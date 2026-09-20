from __future__ import annotations

import csv
import json
from pathlib import Path

from .client import RequestTrace
from .metrics import LevelSummary


def write_reports(
    output_dir: str | Path,
    summaries: list[LevelSummary],
    traces_by_level: dict[int, list[RequestTrace]],
    metadata: dict,
) -> None:
    root = Path(output_dir)
    root.mkdir(parents=True, exist_ok=True)

    with (root / "run.json").open("w", encoding="utf-8") as handle:
        json.dump(metadata, handle, indent=2)

    with (root / "summary.json").open("w", encoding="utf-8") as handle:
        json.dump([summary.as_dict() for summary in summaries], handle, indent=2)

    if summaries:
        with (root / "summary.csv").open(
            "w", newline="", encoding="utf-8"
        ) as handle:
            writer = csv.DictWriter(
                handle, fieldnames=list(summaries[0].as_dict())
            )
            writer.writeheader()
            writer.writerows(summary.as_dict() for summary in summaries)

    with (root / "traces.jsonl").open("w", encoding="utf-8") as handle:
        for concurrency, traces in traces_by_level.items():
            for trace in traces:
                row = {
                    "concurrency": concurrency,
                    "success": trace.success,
                    "ttft_ms": (
                        None if trace.ttft_s is None else trace.ttft_s * 1000
                    ),
                    "inter_chunk_ms": [
                        sample * 1000 for sample in trace.inter_chunk_s
                    ],
                    "tpot_ms": (
                        None if trace.tpot_s is None else trace.tpot_s * 1000
                    ),
                    "e2e_ms": trace.e2e_s * 1000,
                    "prompt_tokens": trace.prompt_tokens,
                    "output_tokens": trace.output_tokens,
                    "error": trace.error,
                }
                handle.write(json.dumps(row) + "\n")
