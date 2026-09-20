package bench

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestStreamRejectsMalformedSSEJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintln(w, "data: {not-json}")
		fmt.Fprintln(w, "data: [DONE]")
	}))
	defer server.Close()

	client := &Client{
		BaseURL: server.URL,
		Model:   "test-model",
		HTTP:    &http.Client{Timeout: time.Second},
	}
	trace := client.Stream(context.Background(), "hello", 8)
	if trace.Success {
		t.Fatalf("expected malformed SSE payload to fail")
	}
	if trace.Error == "" {
		t.Fatalf("expected failure reason to be recorded")
	}
}

func TestStreamCapturesUsageAndTiming(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintln(w, `data: {"choices":[{"delta":{"content":"a"}}]}`)
		fmt.Fprintln(w, `data: {"choices":[{"delta":{"content":"b"}}]}`)
		fmt.Fprintln(w, `data: {"choices":[],"usage":{"prompt_tokens":4,"completion_tokens":2}}`)
		fmt.Fprintln(w, "data: [DONE]")
	}))
	defer server.Close()

	client := &Client{
		BaseURL: server.URL,
		Model:   "test-model",
		HTTP:    &http.Client{Timeout: time.Second},
	}
	trace := client.Stream(context.Background(), "hello", 8)
	if !trace.Success {
		t.Fatalf("unexpected failure: %s", trace.Error)
	}
	if trace.TTFTSeconds == nil {
		t.Fatalf("expected TTFT measurement")
	}
	if len(trace.ITLSeconds) != 1 {
		t.Fatalf("expected one inter-chunk interval, got %d", len(trace.ITLSeconds))
	}
	if trace.PromptTokens == nil || *trace.PromptTokens != 4 {
		t.Fatalf("expected prompt token usage")
	}
	if trace.OutputTokens == nil || *trace.OutputTokens != 2 {
		t.Fatalf("expected output token usage")
	}
}
