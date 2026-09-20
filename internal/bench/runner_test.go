package bench

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRunFailsWhenWarmupRequestFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	_, err := Run(context.Background(), SweepConfig{
		BaseURL:          server.URL,
		Model:            "test-model",
		Concurrency:      []int{1},
		RequestsPerLevel: 1,
		MaxTokens:        8,
		BasePrompt:       "hello",
		Warmups:          1,
		Timeout:          time.Second,
	})
	if err == nil {
		t.Fatalf("expected warmup failure")
	}
	if !strings.Contains(err.Error(), "warm-up request failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}
