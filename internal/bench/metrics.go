package bench

import (
	"math"
	"sort"
)

func Percentile(values []float64, q float64) *float64 {
	if len(values) == 0 {
		return nil
	}
	if q < 0 || q > 100 {
		panic("percentile must be in [0,100]")
	}
	ordered := append([]float64(nil), values...)
	sort.Float64s(ordered)
	if len(ordered) == 1 {
		v := ordered[0]
		return &v
	}
	rank := float64(len(ordered)-1) * q / 100
	lo := int(math.Floor(rank))
	hi := int(math.Ceil(rank))
	v := ordered[lo]
	if lo != hi {
		f := rank - float64(lo)
		v = ordered[lo]*(1-f) + ordered[hi]*f
	}
	return &v
}

func ms(values []float64, q float64) *float64 {
	p := Percentile(values, q)
	if p == nil {
		return nil
	}
	v := *p * 1000
	return &v
}

func Summarize(concurrency int, traces []Trace, durationSeconds float64) LevelSummary {
	var ttft, itl, tpot, e2e []float64
	success := 0
	allPromptCounts := true
	allOutputCounts := true
	promptTotal := 0
	outputTotal := 0

	for _, t := range traces {
		if !t.Success {
			continue
		}
		success++
		e2e = append(e2e, t.E2ESeconds)
		if t.TTFTSeconds != nil {
			ttft = append(ttft, *t.TTFTSeconds)
		}
		itl = append(itl, t.ITLSeconds...)
		if v := t.TPOTSeconds(); v != nil {
			tpot = append(tpot, *v)
		}
		if t.PromptTokens == nil {
			allPromptCounts = false
		} else {
			promptTotal += *t.PromptTokens
		}
		if t.OutputTokens == nil {
			allOutputCounts = false
		} else {
			outputTotal += *t.OutputTokens
		}
	}

	rps := 0.0
	if durationSeconds > 0 {
		rps = float64(success) / durationSeconds
	}
	var outputTPS, totalTPS *float64
	if success > 0 && durationSeconds > 0 && allOutputCounts {
		v := float64(outputTotal) / durationSeconds
		outputTPS = &v
		if allPromptCounts {
			t := float64(promptTotal+outputTotal) / durationSeconds
			totalTPS = &t
		}
	}

	return LevelSummary{
		Concurrency:          concurrency,
		Attempted:            len(traces),
		Succeeded:            success,
		Failed:               len(traces) - success,
		DurationSeconds:      durationSeconds,
		RequestThroughputRPS: rps,
		OutputThroughputTPS:  outputTPS,
		TotalThroughputTPS:   totalTPS,
		TTFTP50MS:            ms(ttft, 50),
		TTFTP90MS:            ms(ttft, 90),
		TTFTP95MS:            ms(ttft, 95),
		TTFTP99MS:            ms(ttft, 99),
		ITLP50MS:             ms(itl, 50),
		ITLP90MS:             ms(itl, 90),
		ITLP95MS:             ms(itl, 95),
		ITLP99MS:             ms(itl, 99),
		TPOTP50MS:            ms(tpot, 50),
		TPOTP99MS:            ms(tpot, 99),
		E2EP50MS:             ms(e2e, 50),
		E2EP99MS:             ms(e2e, 99),
	}
}
