package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/imtisaalm/ttft-itl-benchmark/internal/bench"
)

func parseLevels(raw string) ([]int, error) {
	parts := strings.Split(raw, ",")
	levels := make([]int, 0, len(parts))
	for _, part := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || n <= 0 {
			return nil, fmt.Errorf("invalid concurrency level %q", part)
		}
		levels = append(levels, n)
	}
	if len(levels) == 0 {
		return nil, fmt.Errorf("at least one concurrency level is required")
	}
	return levels, nil
}

func main() {
	url := flag.String("url", "http://127.0.0.1:8000", "OpenAI-compatible base URL")
	model := flag.String("model", "local-model", "served model name")
	concurrencyRaw := flag.String("concurrency", "1,2,4,8,16", "comma-separated concurrency levels")
	requests := flag.Int("requests", 50, "requests per concurrency level")
	maxTokens := flag.Int("max-tokens", 128, "maximum generated tokens per request")
	warmups := flag.Int("warmup", 2, "warm-up requests excluded from results")
	prompt := flag.String("prompt", "Explain the role of a KV cache during autoregressive decoding.", "benchmark prompt")
	sharedPrefix := flag.Bool("shared-prefix", false, "intentionally retain a shared prompt prefix")
	engine := flag.String("engine", "openai-compatible", "runtime label recorded in run metadata")
	hardware := flag.String("hardware", "unspecified", "hardware label recorded in run metadata")
	output := flag.String("output", "results", "result directory")
	timeout := flag.Duration("timeout", 3*time.Minute, "per-request HTTP timeout")
	flag.Parse()

	levels, err := parseLevels(*concurrencyRaw)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	cfg := bench.SweepConfig{
		BaseURL:          *url,
		Model:            *model,
		Concurrency:      levels,
		RequestsPerLevel: *requests,
		MaxTokens:        *maxTokens,
		BasePrompt:       *prompt,
		Warmups:          *warmups,
		SharedPrefix:     *sharedPrefix,
		Timeout:          *timeout,
	}
	result, err := bench.Run(context.Background(), cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	meta := bench.RunMetadata{
		Engine:           *engine,
		Hardware:         *hardware,
		BaseURL:          *url,
		Model:            *model,
		Concurrency:      levels,
		RequestsPerLevel: *requests,
		MaxTokens:        *maxTokens,
		Warmups:          *warmups,
		SharedPrefix:     *sharedPrefix,
	}
	if err := bench.WriteReports(*output, meta, result); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("concurrency  rps    out_tok/s  ttft_p50  ttft_p99  itl_p50  itl_p99")
	for _, row := range result.Summaries {
		fmt.Printf("%11d  %5.2f  %9s  %8s  %8s  %7s  %7s\n",
			row.Concurrency,
			row.RequestThroughputRPS,
			display(row.OutputThroughputTPS),
			display(row.TTFTP50MS),
			display(row.TTFTP99MS),
			display(row.ITLP50MS),
			display(row.ITLP99MS),
		)
	}
}

func display(v *float64) string {
	if v == nil {
		return "n/a"
	}
	return fmt.Sprintf("%.2f", *v)
}
