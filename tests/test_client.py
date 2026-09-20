from ttft_itl_benchmark.client import RequestTrace, parse_sse_data


def test_parse_sse_data() -> None:
    assert parse_sse_data('data: {"choices": []}') == {"choices": []}
    assert parse_sse_data("data: [DONE]") is None


def test_tpot_excludes_first_token_interval() -> None:
    trace = RequestTrace(True, 0.1, (0.02, 0.03), 0.5, 10, 5)
    assert trace.tpot_s == 0.1
