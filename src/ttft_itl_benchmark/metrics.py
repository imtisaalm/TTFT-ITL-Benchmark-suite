from __future__ import annotations

from dataclasses import asdict, dataclass
import math
import statistics

from .client import RequestTrace


def percentile(values: list[float], q: float) -> float | None:
    if not values:
        return None
    if not 0 <= q <= 100:
        raise ValueError("percentile must be in [0, 100]")
    ordered = sorted(values)
    if len(ordered) == 1:
        return ordered[0]
    rank = (len(ordered) - 1) * q / 100
    low = math.floor(rank)
    high = math.ceil(rank)
    if low == high:
        return ordered[low]
    fraction = rank - low
    return ordered[low] * (1 - fraction) + ordered[high] * fraction


@dataclass(frozen=True)
class LevelSummary:
    concurrency: int
    attempted: int
    succeeded: int
    failed: int
    duration_s: float
    request_throughput_rps: float
    output_throughput_tps: float | None
    total_throughput_tps: float | None
    mean_prompt_tokens: float | None
    mean_output_tokens: float | None
    ttft_p50_ms: float | None
    ttft_p90_ms: float | None
    ttft_p95_ms: float | None
    ttft_p99_ms: float | None
    itl_p50_ms: float | None
    itl_p90_ms: float | None
    itl_p95_ms: float | None
    itl_p99_ms: float | None
    tpot_p50_ms: float | None
    tpot_p99_ms: float | None
    e2e_p50_ms: float | None
    e2e_p99_ms: float | None

    def as_dict(self) -> dict:
        return asdict(self)


def _ms(values: list[float], q: float) -> float | None:
    value = percentile(values, q)
    return None if value is None else value * 1000


def summarize_level(
    *, concurrency: int, traces: list[RequestTrace], duration_s: float
) -> LevelSummary:
    successful = [trace for trace in traces if trace.success]
    ttft = [trace.ttft_s for trace in successful if trace.ttft_s is not None]
    itl = [sample for trace in successful for sample in trace.inter_chunk_s]
    tpot = [trace.tpot_s for trace in successful if trace.tpot_s is not None]
    e2e = [trace.e2e_s for trace in successful]
    prompt_counts = [
        trace.prompt_tokens
        for trace in successful
        if trace.prompt_tokens is not None
    ]
    output_counts = [
        trace.output_tokens
        for trace in successful
        if trace.output_tokens is not None
    ]

    output_tps = None
    total_tps = None
    if duration_s > 0 and len(output_counts) == len(successful) and successful:
        output_tps = sum(output_counts) / duration_s
        if len(prompt_counts) == len(successful):
            total_tps = (sum(prompt_counts) + sum(output_counts)) / duration_s

    return LevelSummary(
        concurrency=concurrency,
        attempted=len(traces),
        succeeded=len(successful),
        failed=len(traces) - len(successful),
        duration_s=duration_s,
        request_throughput_rps=0.0 if duration_s <= 0 else len(successful) / duration_s,
        output_throughput_tps=output_tps,
        total_throughput_tps=total_tps,
        mean_prompt_tokens=statistics.fmean(prompt_counts) if prompt_counts else None,
        mean_output_tokens=statistics.fmean(output_counts) if output_counts else None,
        ttft_p50_ms=_ms(ttft, 50),
        ttft_p90_ms=_ms(ttft, 90),
        ttft_p95_ms=_ms(ttft, 95),
        ttft_p99_ms=_ms(ttft, 99),
        itl_p50_ms=_ms(itl, 50),
        itl_p90_ms=_ms(itl, 90),
        itl_p95_ms=_ms(itl, 95),
        itl_p99_ms=_ms(itl, 99),
        tpot_p50_ms=_ms(tpot, 50),
        tpot_p99_ms=_ms(tpot, 99),
        e2e_p50_ms=_ms(e2e, 50),
        e2e_p99_ms=_ms(e2e, 99),
    )
