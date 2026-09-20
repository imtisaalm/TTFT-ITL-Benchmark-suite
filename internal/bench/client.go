package bench

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL string
	Model   string
	APIKey  string
	HTTP    *http.Client
}

type streamRequest struct {
	Model         string        `json:"model"`
	Messages      []chatMessage `json:"messages"`
	Temperature   float64       `json:"temperature"`
	MaxTokens     int           `json:"max_tokens"`
	Stream        bool          `json:"stream"`
	StreamOptions streamOptions `json:"stream_options"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type streamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type chunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage,omitempty"`
}

func (c *Client) Stream(ctx context.Context, prompt string, maxTokens int) Trace {
	body, err := json.Marshal(streamRequest{
		Model:         c.Model,
		Messages:      []chatMessage{{Role: "user", Content: prompt}},
		Temperature:   0,
		MaxTokens:     maxTokens,
		Stream:        true,
		StreamOptions: streamOptions{IncludeUsage: true},
	})
	if err != nil {
		return Trace{Success: false, Error: err.Error()}
	}

	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.BaseURL, "/")+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return Trace{Success: false, Error: err.Error()}
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return Trace{Success: false, E2ESeconds: time.Since(start).Seconds(), Error: err.Error()}
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Trace{Success: false, E2ESeconds: time.Since(start).Seconds(), Error: fmt.Sprintf("HTTP %d", resp.StatusCode)}
	}

	var first *time.Time
	var previous *time.Time
	var intervals []float64
	var promptTokens, outputTokens *int

	scanner := bufio.NewScanner(resp.Body)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 4*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}
		var ch chunk
		if err := json.Unmarshal([]byte(data), &ch); err != nil {
			continue
		}
		if ch.Usage != nil {
			p := ch.Usage.PromptTokens
			o := ch.Usage.CompletionTokens
			promptTokens = &p
			outputTokens = &o
		}
		if len(ch.Choices) == 0 || ch.Choices[0].Delta.Content == "" {
			continue
		}
		now := time.Now()
		if first == nil {
			t := now
			first = &t
		}
		if previous != nil {
			intervals = append(intervals, now.Sub(*previous).Seconds())
		}
		t := now
		previous = &t
	}
	if err := scanner.Err(); err != nil {
		return Trace{Success: false, E2ESeconds: time.Since(start).Seconds(), Error: err.Error()}
	}

	var ttft *float64
	if first != nil {
		v := first.Sub(start).Seconds()
		ttft = &v
	}
	return Trace{
		Success:      true,
		TTFTSeconds:  ttft,
		ITLSeconds:   intervals,
		E2ESeconds:   time.Since(start).Seconds(),
		PromptTokens: promptTokens,
		OutputTokens: outputTokens,
	}
}
