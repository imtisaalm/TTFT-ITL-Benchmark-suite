from __future__ import annotations

import asyncio
from dataclasses import dataclass
import os
import time

import httpx

from .client import RequestTrace, stream_request
from .metrics import LevelSummary, summarize_level
from .workload import prompt_for_request


@dataclass(frozen=True)
class SweepConfig:
    base_url: str
    model: str
    concurrency_levels: tuple[int, ...]
    requests_per_level: int
    max_tokens: int
    base_prompt: str
    warmup_requests: int = 2
    shared_prefix: bool = False
    timeout_s: float = 180.0

    def validate(self) -> None:
        if not self.concurrency_levels or any(level <= 0 for level in self.concurrency_levels):
            raise ValueError("concurrency levels must be positive")
        if self.requests_per_level <= 0 or self.max_tokens <= 0:
            raise ValueError("requests_per_level and max_tokens must be positive")
        if self.warmup_requests < 0:
            raise ValueError("warmup_requests must be non-negative")


async def _one(
    client: httpx.AsyncClient,
    semaphore: asyncio.Semaphore,
    config: SweepConfig,
    request_id: int,
) -> RequestTrace:
    async with semaphore:
        prompt = prompt_for_request(config.base_prompt, request_id, config.shared_prefix)
        return await stream_request(
            client,
            base_url=config.base_url,
            model=config.model,
            prompt=prompt,
            max_tokens=config.max_tokens,
            api_key=os.getenv("OPENAI_API_KEY"),
        )


async def _run_level(
    client: httpx.AsyncClient,
    config: SweepConfig,
    concurrency: int,
    id_offset: int,
) -> tuple[LevelSummary, list[RequestTrace]]:
    semaphore = asyncio.Semaphore(concurrency)
    start = time.perf_counter()
    traces = await asyncio.gather(
        *[
            _one(client, semaphore, config, id_offset + i)
            for i in range(config.requests_per_level)
        ]
    )
    duration = time.perf_counter() - start
    trace_list = list(traces)
    return summarize_level(
        concurrency=concurrency,
        traces=trace_list,
        duration_s=duration,
    ), trace_list


async def run_sweep(
    config: SweepConfig,
) -> tuple[list[LevelSummary], dict[int, list[RequestTrace]]]:
    config.validate()
    timeout = httpx.Timeout(config.timeout_s)
    limits = httpx.Limits(max_connections=max(config.concurrency_levels) + 8)
    summaries: list[LevelSummary] = []
    traces_by_level: dict[int, list[RequestTrace]] = {}

    async with httpx.AsyncClient(timeout=timeout, limits=limits) as client:
        for i in range(config.warmup_requests):
            await stream_request(
                client,
                base_url=config.base_url,
                model=config.model,
                prompt=prompt_for_request(
                    config.base_prompt, -1 - i, config.shared_prefix
                ),
                max_tokens=min(config.max_tokens, 32),
                api_key=os.getenv("OPENAI_API_KEY"),
            )

        offset = 0
        for concurrency in config.concurrency_levels:
            summary, traces = await _run_level(client, config, concurrency, offset)
            summaries.append(summary)
            traces_by_level[concurrency] = traces
            offset += config.requests_per_level

    return summaries, traces_by_level
