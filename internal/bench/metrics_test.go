package bench

import (
	"math"
	"testing"
)

func TestPercentile(t *testing.T) {
	v := Percentile([]float64{1, 2, 3, 4}, 50)
	if v == nil || *v != 2.5 {
		t.Fatalf("expected 2.5, got %v", v)
	}
}

func TestSummarize(t *testing.T) {
	ttft1, ttft2 := 0.1, 0.2
	p1, p2 := 10, 12
	o1, o2 := 3, 3
	traces := []Trace{
		{Success: true, TTFTSeconds: &ttft1, ITLSeconds: []float64{0.02, 0.03}, E2ESeconds: 0.5, PromptTokens: &p1, OutputTokens: &o1},
		{Success: true, TTFTSeconds: &ttft2, ITLSeconds: []float64{0.04, 0.05}, E2ESeconds: 0.6, PromptTokens: &p2, OutputTokens: &o2},
	}
	s := Summarize(2, traces, 1.0)
	if s.OutputThroughputTPS == nil || *s.OutputThroughputTPS != 6 {
		t.Fatalf("expected 6 output tokens/s")
	}
	if s.TTFTP50MS == nil || math.Abs(*s.TTFTP50MS-150) > 1e-9 {
		t.Fatalf("expected TTFT p50 150 ms")
	}
}
