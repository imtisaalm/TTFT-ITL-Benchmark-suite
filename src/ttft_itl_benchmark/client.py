from __future__ import annotations

from dataclasses import dataclass
import json
import time

import httpx


@dataclass(frozen=True)
class RequestTrace:
    success: bool
    ttft_s: float | None
    inter_chunk_s: tuple[float, ...]
    e2e_s: float
    prompt_tokens: int | None
    output_tokens: int | None
    error: str | None = None

    @property
    def tpot_s(self) -> float | None:
        if self.ttft_s is None or self.output_tokens is None or self.output_tokens <= 1:
            return None
        return max(0.0, self.e2e_s - self.ttft_s) / (self.output_tokens - 1)


def parse_sse_data(line: str) -> dict | None:
    if not line.startswith("data: "):
        return None
    data = line[6:]
    if data == "[DONE]":
        return None
    return json.loads(data)


async def stream_request(
    client: httpx.AsyncClient,
    *,
    base_url: str,
    model: str,
    prompt: str,
    max_tokens: int,
    api_key: str | None = None,
) -> RequestTrace:
    headers = {"Content-Type": "application/json"}
    if api_key:
        headers["Authorization"] = f"Bearer {api_key}"

    payload = {
        "model": model,
        "messages": [{"role": "user", "content": prompt}],
        "temperature": 0,
        "max_tokens": max_tokens,
        "stream": True,
        "stream_options": {"include_usage": True},
    }

    start = time.perf_counter()
    first_content_at: float | None = None
    previous_content_at: float | None = None
    intervals: list[float] = []
    prompt_tokens: int | None = None
    output_tokens: int | None = None

    try:
        async with client.stream(
            "POST",
            f"{base_url.rstrip('/')}/v1/chat/completions",
            headers=headers,
            json=payload,
        ) as response:
            response.raise_for_status()
            async for line in response.aiter_lines():
                parsed = parse_sse_data(line)
                if parsed is None:
                    continue

                usage = parsed.get("usage")
                if usage:
                    prompt_tokens = usage.get("prompt_tokens", prompt_tokens)
                    output_tokens = usage.get("completion_tokens", output_tokens)

                choices = parsed.get("choices") or []
                if not choices:
                    continue
                delta = choices[0].get("delta") or {}
                content = delta.get("content") or ""
                if not content:
                    continue

                now = time.perf_counter()
                if first_content_at is None:
                    first_content_at = now
                if previous_content_at is not None:
                    intervals.append(now - previous_content_at)
                previous_content_at = now

        end = time.perf_counter()
        return RequestTrace(
            success=True,
            ttft_s=None if first_content_at is None else first_content_at - start,
            inter_chunk_s=tuple(intervals),
            e2e_s=end - start,
            prompt_tokens=prompt_tokens,
            output_tokens=output_tokens,
        )
    except Exception as exc:
        end = time.perf_counter()
        return RequestTrace(
            success=False,
            ttft_s=None,
            inter_chunk_s=(),
            e2e_s=end - start,
            prompt_tokens=None,
            output_tokens=None,
            error=f"{type(exc).__name__}: {exc}",
        )
