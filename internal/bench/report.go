package bench

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type RunMetadata struct {
	Engine           string `json:"engine"`
	Hardware         string `json:"hardware"`
	BaseURL          string `json:"base_url"`
	Model            string `json:"model"`
	Concurrency      []int  `json:"concurrency_levels"`
	RequestsPerLevel int    `json:"requests_per_level"`
	MaxTokens        int    `json:"max_tokens"`
	Warmups          int    `json:"warmups"`
	SharedPrefix     bool   `json:"shared_prefix"`
}

func WriteReports(dir string, meta RunMetadata, result SweepResult) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(dir, "run.json"), meta); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(dir, "summary.json"), result.Summaries); err != nil {
		return err
	}
	if err := writeSummaryCSV(filepath.Join(dir, "summary.csv"), result.Summaries); err != nil {
		return err
	}
	return writeTraces(filepath.Join(dir, "traces.jsonl"), result.Traces)
}

func writeJSON(path string, v any) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func writeSummaryCSV(path string, rows []LevelSummary) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	if err := w.Write([]string{"concurrency","attempted","succeeded","failed","duration_s","request_throughput_rps","output_throughput_tps","ttft_p50_ms","ttft_p99_ms","itl_p50_ms","itl_p99_ms"}); err != nil {
		return err
	}
	for _, r := range rows {
		if err := w.Write([]string{
			strconv.Itoa(r.Concurrency),
			strconv.Itoa(r.Attempted),
			strconv.Itoa(r.Succeeded),
			strconv.Itoa(r.Failed),
			fmt.Sprintf("%.6f", r.DurationSeconds),
			fmt.Sprintf("%.6f", r.RequestThroughputRPS),
			formatPtr(r.OutputThroughputTPS),
			formatPtr(r.TTFTP50MS),
			formatPtr(r.TTFTP99MS),
			formatPtr(r.ITLP50MS),
			formatPtr(r.ITLP99MS),
		}); err != nil {
			return err
		}
	}
	return w.Error()
}

func writeTraces(path string, traces map[int][]Trace) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	enc := json.NewEncoder(w)
	for concurrency, rows := range traces {
		for _, t := range rows {
			record := struct {
				Concurrency int   `json:"concurrency"`
				Trace       Trace `json:"trace"`
			}{concurrency, t}
			if err := enc.Encode(record); err != nil {
				return err
			}
		}
	}
	return nil
}

func formatPtr(v *float64) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%.6f", *v)
}
