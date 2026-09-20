from ttft_itl_benchmark.workload import prompt_for_request


def test_unique_prefix_is_default() -> None:
    prompt = prompt_for_request("base", 7)
    assert prompt.startswith("[benchmark-request=7]")


def test_shared_prefix_mode_moves_marker_to_end() -> None:
    prompt = prompt_for_request("base", 7, shared_prefix=True)
    assert prompt.startswith("base")
    assert prompt.endswith("[benchmark-request=7]")
