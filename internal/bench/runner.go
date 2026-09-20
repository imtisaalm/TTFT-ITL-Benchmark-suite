package bench

import (
	"context"
	"net/http"
	"os"
	"sync"
	"time"
)

type SweepConfig struct {
	BaseURL          string
	Model            string
	Concurrency      []int
	RequestsPerLevel int
	MaxTokens        int
	BasePrompt       string
	Warmups          int
	SharedPrefix     bool
	Timeout          time.Duration
}

type SweepResult struct {
	Summaries []LevelSummary
	Traces    map[int][]Trace
}

func (c SweepConfig) Validate() error {
	if len(c.Concurrency) == 0 {
		return fmtError("at least one concurrency level is required")
	}
	for _, level := range c.Concurrency {
		if level <= 0 {
			return fmtError("concurrency levels must be positive")
		}
	}
	if c.RequestsPerLevel <= 0 || c.MaxTokens <= 0 {
		return fmtError("requests per level and max tokens must be positive")
	}
	if c.Warmups < 0 {
		return fmtError("warmups must be non-negative")
	}
	return nil
}

type configError string

func (e configError) Error() string { return string(e) }
func fmtError(s string) error       { return configError(s) }

func Run(ctx context.Context, cfg SweepConfig) (SweepResult, error) {
	if err := cfg.Validate(); err != nil {
		return SweepResult{}, err
	}
	httpClient := &http.Client{
		Timeout: cfg.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        256,
			MaxIdleConnsPerHost: 256,
			MaxConnsPerHost:     0,
		},
	}
	defer httpClient.CloseIdleConnections()
	client := &Client{
		BaseURL: cfg.BaseURL,
		Model:   cfg.Model,
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		HTTP:    httpClient,
	}

	for i := 0; i < cfg.Warmups; i++ {
		trace := client.Stream(
			ctx,
			Prompt(cfg.BasePrompt, -1-i, cfg.SharedPrefix),
			min(cfg.MaxTokens, 32),
		)
		if !trace.Success {
			return SweepResult{}, fmtError("warm-up request failed: " + trace.Error)
		}
	}

	result := SweepResult{Traces: make(map[int][]Trace)}
	offset := 0
	for _, level := range cfg.Concurrency {
		start := time.Now()
		traces := runLevel(ctx, client, cfg, level, offset)
		duration := time.Since(start).Seconds()
		result.Traces[level] = traces
		result.Summaries = append(result.Summaries, Summarize(level, traces, duration))
		offset += cfg.RequestsPerLevel
	}
	return result, nil
}

func runLevel(
	ctx context.Context,
	client *Client,
	cfg SweepConfig,
	concurrency int,
	offset int,
) []Trace {
	jobs := make(chan int)
	results := make(chan Trace, cfg.RequestsPerLevel)
	var wg sync.WaitGroup

	workers := min(concurrency, cfg.RequestsPerLevel)
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range jobs {
				prompt := Prompt(cfg.BasePrompt, id, cfg.SharedPrefix)
				results <- client.Stream(ctx, prompt, cfg.MaxTokens)
			}
		}()
	}

	go func() {
		for i := 0; i < cfg.RequestsPerLevel; i++ {
			jobs <- offset + i
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	traces := make([]Trace, 0, cfg.RequestsPerLevel)
	for trace := range results {
		traces = append(traces, trace)
	}
	return traces
}
