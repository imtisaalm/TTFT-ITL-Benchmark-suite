from __future__ import annotations

from dataclasses import asdict, dataclass
from datetime import datetime, timezone
import platform

from .runner import SweepConfig


@dataclass(frozen=True)
class RunMetadata:
    timestamp_utc: str
    engine: str
    hardware: str
    platform: str
    machine: str
    python: str
    base_url: str
    model: str
    concurrency_levels: tuple[int, ...]
    requests_per_level: int
    max_tokens: int
    warmup_requests: int
    shared_prefix: bool

    def as_dict(self) -> dict:
        return asdict(self)


def build_run_metadata(
    config: SweepConfig,
    *,
    engine: str,
    hardware: str,
) -> RunMetadata:
    return RunMetadata(
        timestamp_utc=datetime.now(timezone.utc).isoformat(),
        engine=engine,
        hardware=hardware,
        platform=platform.platform(),
        machine=platform.machine(),
        python=platform.python_version(),
        base_url=config.base_url,
        model=config.model,
        concurrency_levels=config.concurrency_levels,
        requests_per_level=config.requests_per_level,
        max_tokens=config.max_tokens,
        warmup_requests=config.warmup_requests,
        shared_prefix=config.shared_prefix,
    )
