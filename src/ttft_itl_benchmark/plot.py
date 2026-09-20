from __future__ import annotations

import json
from pathlib import Path


def plot_summary(summary_json: str | Path, output_dir: str | Path) -> None:
    import matplotlib.pyplot as plt

    with Path(summary_json).open("r", encoding="utf-8") as handle:
        rows = json.load(handle)

    output = Path(output_dir)
    output.mkdir(parents=True, exist_ok=True)
    concurrency = [row["concurrency"] for row in rows]

    fig, ax = plt.subplots()
    ax.plot(
        concurrency,
        [row["ttft_p50_ms"] for row in rows],
        marker="o",
        label="TTFT p50",
    )
    ax.plot(
        concurrency,
        [row["ttft_p99_ms"] for row in rows],
        marker="o",
        label="TTFT p99",
    )
    ax.set_xlabel("Concurrency")
    ax.set_ylabel("TTFT (ms)")
    ax.set_title("Time to first token under load")
    ax.legend()
    fig.tight_layout()
    fig.savefig(output / "ttft_load_curve.png", dpi=160)
    plt.close(fig)

    fig, ax = plt.subplots()
    ax.plot(
        concurrency,
        [row["output_throughput_tps"] for row in rows],
        marker="o",
        label="Output tokens/s",
    )
    ax.set_xlabel("Concurrency")
    ax.set_ylabel("Output throughput (tokens/s)")
    ax.set_title("Output throughput under load")
    ax.legend()
    fig.tight_layout()
    fig.savefig(output / "output_throughput_curve.png", dpi=160)
    plt.close(fig)
