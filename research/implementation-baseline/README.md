# tradingview-mcp-go implementation baseline

Date: 2026-04-27
OS: Windows
Timezone: Asia/Irkutsk
Scope: factual baseline before implementation research changes. No source code changes were made by this baseline step.

## Summary

- `go test ./...` passed with exit code 0.
- MCP stdio smoke passed for `initialize` and `tools/list`.
- Actual `tools/list` count is 82 tools.
- Expected count in the task was 78 tools, so the 78-tool criterion is not confirmed by the current repository state.
- CLI smoke commands all exited with code 0 and produced valid JSON on stdout.
- TradingView Desktop was reachable through CDP on port 9222.
- TradingView Desktop version inferred from install path: `3.1.0.7818`.
- Active chart during baseline: `RUS:NG1!`, timeframe `1D`, 2 studies.
- `quote` returned `bid: 0` and `ask: 0` for `RUS:NG1!`; this remains consistent with the known MOEX/futures quote issue.

## Artifacts

Environment:

- `environment.json`
- `git-status-after.txt`

Tests:

- `go-test.result.json`
- `go-test.stdout.txt`
- `go-test.stderr.txt`

MCP stdio:

- `mcp-initialize-tools-list.input.jsonl`
- `mcp-initialize-tools-list.output.jsonl`
- `mcp-initialize-tools-list.result.json`
- `mcp-initialize.response.json`
- `mcp-tools-list.response.json`
- `mcp-tools-count.json`
- `mcp-tools-names.json`
- `mcp-initialize-tools-list.stderr.txt`

CLI:

- `cli-summary.json`
- `cli-status.stdout.json`
- `cli-doctor.stdout.json`
- `cli-discover.stdout.json`
- `cli-chart-state.stdout.json`
- `cli-quote.stdout.json`
- `cli-ohlcv.stdout.json`
- `cli-screenshot.stdout.json`
- matching `cli-*.result.json` and `cli-*.stderr.txt` files

Screenshot:

- Tool output path: `screenshots/implementation-baseline-20260427-101853.png`
- Baseline copy: `implementation-baseline-20260427-101853.png`
- Copy metadata: `cli-screenshot-copy.json`

## Commands Run

| Check | Command | Exit | Result |
| --- | --- | ---: | --- |
| Tests | `go test ./...` | 0 | pass |
| MCP stdio | `go run ./cmd/tvmcp < mcp-initialize-tools-list.input.jsonl` | 0 | initialize ok, tools/list ok |
| CLI status | `go run ./cmd/tv status` | 0 | JSON ok, CDP connected |
| CLI doctor | `go run ./cmd/tv doctor` | 0 | JSON ok, TradingView running with CDP flag |
| CLI discover | `go run ./cmd/tv discover` | 0 | JSON ok |
| CLI chart-state | `go run ./cmd/tv chart-state` | 0 | JSON ok |
| CLI quote | `go run ./cmd/tv quote` | 0 | JSON ok |
| CLI ohlcv | `go run ./cmd/tv ohlcv --count 10 --summary` | 0 | JSON ok |
| CLI screenshot | `go run ./cmd/tv screenshot --region chart --filename implementation-baseline-20260427-101853.png` | 0 | JSON ok, PNG saved |

## MCP Tool Count

Result file: `mcp-tools-count.json`

```json
{
  "count": 82,
  "expected": 78,
  "matchesExpected": false
}
```

This baseline records the actual current state. The repository currently exposes 82 MCP tools via `tools/list`; it does not satisfy the requested 78-tool confirmation without code changes.

## Notes

- The MCP server wrote the project warning banner to stderr, not stdout.
- Normal CLI stderr files are empty for all successful CLI smoke commands.
- The screenshot command saved through the normal `screenshots/` directory; a copy is included in this baseline folder for audit convenience.
