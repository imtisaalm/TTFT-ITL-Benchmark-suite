# TTFT / ITL Benchmark Suite

A concurrency-sweep load-testing harness for streamed OpenAI-compatible inference endpoints.

The suite records request-level traces and reports time to first token (TTFT), inter-token/inter-chunk latency (ITL), time per output token (TPOT), end-to-end latency, request throughput, and token throughput as concurrency increases.

## Measurement model

Each concurrency level is run as a closed-loop workload. At most `N` HTTP requests are active at once. When one request completes, the next waiting task enters the client slot. Request timing starts only after the concurrency semaphore is acquired, so client-side semaphore wait is excluded from TTFT and end-to-end latency.

Warm-up requests are issued before the measured sweep and are not included in reported statistics.

### Metrics

**TTFT**

Time from initiating the HTTP request to the first non-empty streamed content chunk.

**ITL**

Intervals between consecutive non-empty streamed content chunks. For servers that emit one token per content-bearing SSE chunk, this is token-level ITL. If a server aggregates multiple tokens into a chunk, the value is inter-chunk latency and should not be interpreted as exact token-level ITL.

**TPOT**

When the endpoint returns streamed usage metadata, TPOT is computed per request as:

```text
(E2E latency - TTFT) / (completion_tokens - 1)
```

**Request throughput**

Successful requests divided by wall-clock duration of the concurrency level.

**Output throughput**

Server-reported completion tokens divided by wall-clock duration. If streamed usage metadata is unavailable, token-throughput fields remain null rather than being estimated from text length.

Percentiles are reported at p50, p90, p95, and p99 for TTFT and ITL, with p50/p99 for TPOT and end-to-end latency.

## Prefix-cache control

By default, a deterministic request identifier is inserted at the beginning of each prompt. This reduces unintended reuse of a long common prefix across repeated benchmark requests.

Use `--shared-prefix` only when prefix-cache reuse is intentionally part of the experiment.

## Installation

```bash
python -m venv .venv
source .venv/bin/activate
python -m pip install -e '.[test]'
```

Optional plotting:

```bash
python -m pip install -e '.[plot]'
```

## Run a concurrency sweep

The command below targets the companion self-hosted vLLM server configuration, but any compatible streaming endpoint can be used.

```bash
ttft-itl-benchmark run \
  --url http://127.0.0.1:8000 \
  --model local-model \
  --concurrency 1,2,4,8,16,32 \
  --requests 50 \
  --max-tokens 128 \
  --engine vllm \
  --hardware "RTX 4090 24GB" \
  --output results/run-001
```

If the endpoint requires authentication, set `OPENAI_API_KEY` in the environment.

The terminal summary is intentionally compact:

```text
concurrency  rps    out_tok/s  ttft_p50  ttft_p99  itl_p50  itl_p99
```

No example performance numbers are checked into the repository because latency and throughput are hardware-, model-, runtime-, and configuration-dependent.

## Result files

Each run directory contains:

```text
run.json        experimental context and host metadata
summary.json    one aggregate record per concurrency level
summary.csv     tabular form of the same aggregate records
traces.jsonl    request-level traces and failure details
```

`run.json` records:

- UTC timestamp
- engine label
- hardware label
- host platform and machine architecture
- Python version
- endpoint and model name
- concurrency levels
- requests per level
- maximum generated tokens
- warm-up count
- prefix-cache mode

This keeps benchmark outputs interpretable when results are compared later.

## Plot load curves

```bash
ttft-itl-benchmark plot \
  results/run-001/summary.json \
  --output results/run-001
```

The plotting command writes separate TTFT and output-throughput curves. Raw JSON/CSV remains the primary result format.

## Tests

```bash
pytest -q
```

The unit tests cover SSE parsing, TPOT calculation, percentile interpolation, throughput aggregation, and prefix-cache workload construction. Network benchmarking is excluded from CI.

## Interpretation

A useful serving curve normally shows two regimes: a region where throughput increases with concurrency at acceptable latency, followed by a saturation region where queueing and TTFT rise faster than throughput. This repository reports the measurements needed to locate that transition; it does not assign a universal capacity threshold.

Comparisons should hold constant at least the model, runtime version, accelerator, quantization/precision, context distribution, output-token limit, scheduler configuration, and prefix-cache mode.

## References

- vLLM benchmark CLI and metric definitions: https://docs.vllm.ai/en/latest/benchmarking/cli/
- vLLM serving metrics: https://docs.vllm.ai/en/latest/design/metrics/
