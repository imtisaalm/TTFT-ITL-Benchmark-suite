# TTFT / ITL Benchmark Suite

A Go load-testing harness for streamed OpenAI-compatible inference endpoints.

The benchmark runs controlled concurrency sweeps and records time to first token (TTFT), inter-token/inter-chunk latency (ITL), time per output token (TPOT), end-to-end latency, request throughput, and token throughput.

## Why Go

The load generator is implemented in Go so client-side concurrency, connection reuse, streaming, and timing are handled without a Python async runtime. Each concurrency level uses a fixed worker pool of goroutines over a shared HTTP transport.

The inference server itself may be implemented in any language or runtime. The only requirement is an OpenAI-compatible streaming chat-completions endpoint.

## Measurement model

Each concurrency level is a closed-loop workload. At most N requests are active at once. When one request completes, a worker accepts the next queued request. Per-request timing starts when the HTTP request is constructed by an active worker, so time spent waiting in the local job queue is excluded.

Warm-up requests run before the measured sweep and are excluded from statistics.

### TTFT

Elapsed time from request initiation to the first non-empty streamed content chunk.

### ITL

Elapsed time between consecutive non-empty streamed content chunks.

When the server emits one token per content-bearing SSE chunk, this is token-level ITL. If a server coalesces several tokens into one chunk, the value is inter-chunk latency and should not be presented as exact token-level ITL.

### TPOT

When completion-token counts are returned through streamed usage metadata:

```text
(E2E latency - TTFT) / (completion_tokens - 1)
```

### Throughput

Request throughput is successful requests divided by level wall time.

Output-token throughput uses the server-reported completion-token count. If the endpoint does not return streamed usage metadata, token-throughput fields remain absent rather than being estimated from text length.

## Build

Requires Go 1.23 or newer.

```bash
go build ./cmd/ttft-itl-benchmark
go test ./...
```

## Run

```bash
go run ./cmd/ttft-itl-benchmark \
  -url http://127.0.0.1:8000 \
  -model local-model \
  -concurrency 1,2,4,8,16,32 \
  -requests 50 \
  -max-tokens 128 \
  -engine vllm \
  -hardware "RTX 4090 24GB" \
  -output results/run-001
```

Set `OPENAI_API_KEY` when authentication is enabled.

The terminal output is intentionally compact:

```text
concurrency  rps    out_tok/s  ttft_p50  ttft_p99  itl_p50  itl_p99
```

No synthetic performance results are committed. Latency and throughput depend on the model, serving engine, accelerator, precision, context distribution, scheduler configuration, and cache state.

## Prefix-cache control

By default, a deterministic request identifier is inserted at the beginning of the prompt. This prevents the benchmark from unintentionally measuring a long shared prefix from an enabled prefix cache.

Use `-shared-prefix` only when prefix reuse is intentionally part of the experiment.

## Output

Each run writes:

```text
run.json        experiment and host metadata
summary.json    aggregate metrics by concurrency
summary.csv     tabular aggregate metrics
traces.jsonl    individual request traces
```

`run.json` includes:

- UTC timestamp
- Go version
- operating system and architecture
- serving-engine label
- hardware label
- endpoint and model
- concurrency sweep
- request count
- output-token limit
- warm-up count
- prefix-cache mode

## Implementation

```text
cmd/ttft-itl-benchmark/
    main.go             CLI

internal/bench/
    client.go           SSE streaming client
    runner.go           goroutine worker pools and load sweep
    metrics.go          percentile and throughput aggregation
    workload.go         deterministic prompt construction
    report.go           JSON, CSV, and trace persistence
```

The HTTP transport is shared across workers so connections can be reused. The scanner buffer is increased for streamed responses, and individual request failures are recorded in the trace set instead of terminating an entire sweep.

## Tests

```bash
go test ./...
```

CI runs `go vet` and the full Go test suite.

## Interpretation

The useful operating region is generally the range in which throughput continues to rise without disproportionate growth in queueing and TTFT. The tool reports the load curve; it does not define a universal saturation threshold.

Comparisons should hold constant the model, engine version, accelerator, precision or quantization, prompt/output distribution, scheduler configuration, and prefix-cache policy.

## References

- vLLM benchmark terminology: https://docs.vllm.ai/en/latest/benchmarking/cli/
- vLLM serving metrics: https://docs.vllm.ai/en/latest/design/metrics/
