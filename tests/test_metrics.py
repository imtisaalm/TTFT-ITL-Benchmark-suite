import pytest

from ttft_itl_benchmark.client import RequestTrace
from ttft_itl_benchmark.metrics import percentile, summarize_level


def test_percentile_interpolates() -> None:
    assert percentile([1.0, 2.0, 3.0, 4.0], 50) == 2.5


def test_level_summary_uses_server_token_counts() -> None:
    traces = [
        RequestTrace(True, 0.1, (0.02, 0.03), 0.5, 10, 3),
        RequestTrace(True, 0.2, (0.04, 0.05), 0.6, 12, 3),
    ]
    summary = summarize_level(
        concurrency=2,
        traces=traces,
        duration_s=1.0,
    )
    assert summary.request_throughput_rps == 2.0
    assert summary.output_throughput_tps == 6.0
    assert summary.total_throughput_tps == 28.0
    assert summary.ttft_p50_ms == pytest.approx(150.0)
