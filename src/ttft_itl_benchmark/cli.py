from __future__ import annotations

import asyncio
from pathlib import Path

import typer

from .metadata import build_run_metadata
from .plot import plot_summary
from .report import write_reports
from .runner import SweepConfig, run_sweep

app = typer.Typer(add_completion=False, no_args_is_help=True)


def _levels(raw: str) -> tuple[int, ...]:
    values = tuple(
        int(part.strip())
        for part in raw.split(",")
        if part.strip()
    )
    if not values or any(value <= 0 for value in values):
        raise typer.BadParameter(
            "concurrency must be a comma-separated list of positive integers"
        )
    return values


def _fmt(value: float | None) -> str:
    return "n/a" if value is None else f"{value:.2f}"


@app.command("run")
def run_command(
    url: str = typer.Option("http://127.0.0.1:8000"),
    model: str = typer.Option("local-model"),
    concurrency: str = typer.Option("1,2,4,8,16"),
    requests: int = typer.Option(50, min=1),
    max_tokens: int = typer.Option(128, min=1),
    warmup: int = typer.Option(2, min=0),
    prompt: str = typer.Option(
        "Explain the role of a KV cache during autoregressive decoding."
    ),
    shared_prefix: bool = typer.Option(
        False,
        help="Intentionally retain a shared prompt prefix.",
    ),
    engine: str = typer.Option("openai-compatible"),
    hardware: str = typer.Option("unspecified"),
    output: Path = typer.Option(Path("results")),
) -> None:
    """Run a closed-loop concurrency sweep."""
    config = SweepConfig(
        base_url=url,
        model=model,
        concurrency_levels=_levels(concurrency),
        requests_per_level=requests,
        max_tokens=max_tokens,
        base_prompt=prompt,
        warmup_requests=warmup,
        shared_prefix=shared_prefix,
    )
    summaries, traces = asyncio.run(run_sweep(config))
    metadata = build_run_metadata(
        config,
        engine=engine,
        hardware=hardware,
    )
    write_reports(output, summaries, traces, metadata.as_dict())

    typer.echo(
        "concurrency  rps    out_tok/s  "
        "ttft_p50  ttft_p99  itl_p50  itl_p99"
    )
    for row in summaries:
        typer.echo(
            f"{row.concurrency:>11}  "
            f"{_fmt(row.request_throughput_rps):>5}  "
            f"{_fmt(row.output_throughput_tps):>9}  "
            f"{_fmt(row.ttft_p50_ms):>8}  "
            f"{_fmt(row.ttft_p99_ms):>8}  "
            f"{_fmt(row.itl_p50_ms):>7}  "
            f"{_fmt(row.itl_p99_ms):>7}"
        )


@app.command()
def plot(
    summary: Path = typer.Argument(..., exists=True, readable=True),
    output: Path = typer.Option(Path("results")),
) -> None:
    """Generate TTFT and throughput load curves from summary.json."""
    plot_summary(summary, output)


if __name__ == "__main__":
    app()
