package bench

type Trace struct {
	Success      bool      `json:"success"`
	TTFTSeconds  *float64  `json:"ttft_seconds,omitempty"`
	ITLSeconds   []float64 `json:"itl_seconds"`
	E2ESeconds   float64   `json:"e2e_seconds"`
	PromptTokens *int      `json:"prompt_tokens,omitempty"`
	OutputTokens *int      `json:"output_tokens,omitempty"`
	Error        string    `json:"error,omitempty"`
}

func (t Trace) TPOTSeconds() *float64 {
	if t.TTFTSeconds == nil || t.OutputTokens == nil || *t.OutputTokens <= 1 {
		return nil
	}
	v := (t.E2ESeconds - *t.TTFTSeconds) / float64(*t.OutputTokens-1)
	if v < 0 {
		v = 0
	}
	return &v
}

type LevelSummary struct {
	Concurrency          int      `json:"concurrency"`
	Attempted            int      `json:"attempted"`
	Succeeded            int      `json:"succeeded"`
	Failed               int      `json:"failed"`
	DurationSeconds      float64  `json:"duration_s"`
	RequestThroughputRPS float64  `json:"request_throughput_rps"`
	OutputThroughputTPS  *float64 `json:"output_throughput_tps,omitempty"`
	TotalThroughputTPS   *float64 `json:"total_throughput_tps,omitempty"`
	TTFTP50MS            *float64 `json:"ttft_p50_ms,omitempty"`
	TTFTP90MS            *float64 `json:"ttft_p90_ms,omitempty"`
	TTFTP95MS            *float64 `json:"ttft_p95_ms,omitempty"`
	TTFTP99MS            *float64 `json:"ttft_p99_ms,omitempty"`
	ITLP50MS             *float64 `json:"itl_p50_ms,omitempty"`
	ITLP90MS             *float64 `json:"itl_p90_ms,omitempty"`
	ITLP95MS             *float64 `json:"itl_p95_ms,omitempty"`
	ITLP99MS             *float64 `json:"itl_p99_ms,omitempty"`
	TPOTP50MS            *float64 `json:"tpot_p50_ms,omitempty"`
	TPOTP99MS            *float64 `json:"tpot_p99_ms,omitempty"`
	E2EP50MS             *float64 `json:"e2e_p50_ms,omitempty"`
	E2EP99MS             *float64 `json:"e2e_p99_ms,omitempty"`
}
