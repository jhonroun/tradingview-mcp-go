# Study Limit Detection

Date: 2026-04-27
OS: Windows
TradingView access: local TradingView Desktop through CDP
Symbol during smoke: `RUS:NG1!`

## Implemented behavior

- `chart_manage_indicator` no longer returns silent success when `createStudy`
  does not produce a new study entity.
- TradingView study-limit messages are classified as:

```json
{
  "success": false,
  "status": "study_limit_reached",
  "currentStudies": [],
  "limit": 2,
  "suggestion": "Remove one study manually, upgrade the TradingView plan, or retry with allow_remove_any=true to let this tool remove the most recent study and retry."
}
```

- `allow_remove_any` is false by default.
- With `allow_remove_any=true`, `chart_manage_indicator` may remove only the
  most recent existing study, write a JSONL entry to
  `research/study-limit-detection/removals.jsonl`, and retry `createStudy`
  once.
- No study is removed automatically without the explicit flag.
- CLI equivalent:

```powershell
go run ./cmd/tv manage-indicator add "Moving Average" --allow-remove-any
go run ./cmd/tv manage-indicator remove ENTITY_ID
```

## Live smoke

The current TradingView session started with 2 studies:

- `Vvzmzg` / `Помошник RSI - True ADX`
- `cfyrsD` / `Volume`

Attempt:

```powershell
go run ./cmd/tv manage-indicator add "Moving Average"
```

Result: the current account/session did not hit the Basic 2-study limit.
TradingView added `Moving Average` as entity `5C13zY`.

Cleanup:

```powershell
go run ./cmd/tv manage-indicator remove 5C13zY
```

After cleanup, the chart returned to the original 2 studies.

## Artifacts

- `chart-state.before.json`
- `cli-add-moving-average.no-allow.stdout.json`
- `cli-add-moving-average.no-allow.stderr.txt`
- `cli-add-moving-average.no-allow.exitcode.txt`
- `cli-remove-moving-average.cleanup.stdout.json`
- `cli-remove-moving-average.cleanup.stderr.txt`
- `cli-remove-moving-average.cleanup.exitcode.txt`
- `chart-state.after-cleanup.json`

## Verification

- Unit tests cover English and Russian limit-message parsing.
- Unit tests verify that non-limit add failures do not return `success:true`.
- Unit tests verify JSONL removal-log writing through a temp path.
- `go test ./...` passed.
- `go vet ./...` passed.
