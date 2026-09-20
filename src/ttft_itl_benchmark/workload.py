from __future__ import annotations


def prompt_for_request(base_prompt: str, request_id: int, shared_prefix: bool = False) -> str:
    """Generate deterministic prompts for a concurrency sweep.

    By default the request identifier is placed at the beginning so repeated runs
    do not accidentally benchmark a long shared prefix from the prefix cache.
    Set shared_prefix when prefix reuse is intentionally part of the experiment.
    """
    marker = f"[benchmark-request={request_id}]"
    if shared_prefix:
        return f"{base_prompt}\n{marker}"
    return f"{marker}\n{base_prompt}"
