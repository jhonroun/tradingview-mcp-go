# CDP evaluate awaitPromise baseline

Date: 2026-04-27
OS: Windows
TradingView Desktop: reachable through CDP on port 9222

## Summary

- Added an internal CDP `EvaluateWithOptions` helper.
- The helper supports `awaitPromise`, `returnByValue`, and `timeout`.
- Existing `Client.Evaluate(ctx, expression, awaitPromise)` remains available and keeps `returnByValue: true`.
- Existing MCP `ui_evaluate` public API was not changed.
- CLI-only `tv ui eval-await` was used as the live async smoke path.

## Live Smoke

| Check | Command | Exit | Result |
| --- | --- | ---: | --- |
| Existing sync evaluate path | `go run ./cmd/tv ui eval 21+21` | 0 | `result: 42` |
| Async awaitPromise path | `go run ./cmd/tv ui eval-await "(async () => 42)()"` | 0 | `result: 42` |

## Artifacts

- `smoke-summary.json`
- `ui-eval-sync.stdout.json`
- `ui-eval-sync.stderr.txt`
- `ui-eval-sync.result.json`
- `ui-eval-await-async.stdout.json`
- `ui-eval-await-async.stderr.txt`
- `ui-eval-await-async.result.json`

## Tests

- `go test ./internal/cdp ./internal/tools/ui`
- `go test ./...`
- `go vet ./...`
