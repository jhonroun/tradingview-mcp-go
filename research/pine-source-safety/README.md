# Pine Source Safety

Date: 2026-04-27
OS: Windows

## Safe Workflow

1. Read the current editor state:

   ```text
   tv pine get
   ```

   The response includes `source`, `source_sha256`, `hash`, `script_name`,
   `script_type`, `pine_version`, `line_count`, and `char_count`.

2. Backup before mutation:

   `pine_set_source`, `tv pine set`, `tv pine set-file`, `pine_new`, and
   `pine_open` create a backup before writing into Monaco.

   Backup files are written under:

   ```text
   research/pine-source-safety/session-*/backup.json
   research/pine-source-safety/session-*/*.pine
   ```

3. Set source with an optional hash guard:

   ```text
   tv pine set-file PATH --expected-current-sha256 HASH
   ```

   If the editor hash differs from `expected-current-sha256`, the write is
   refused.

4. Compile and read structured diagnostics:

   ```text
   tv pine compile
   tv pine errors
   ```

   Compile/error responses include `compiled`, `has_errors`, `error_count`,
   `warning_count`, `errors`, `warnings`, and `diagnostics`.

5. Verify chart/Pine state with the relevant read tools.

6. Restore from backup:

   ```text
   tv pine restore research/pine-source-safety/session-.../backup.json
   ```

   Restore verifies SHA256 before writing and again after writing. `.pine`
   restore requires an explicit `--expected-sha256 HASH`.

## Live Smoke

Files:

- `live-pine-get-source.json` - current editor source and metadata.
- `live-pine-set-same-source.json` - wrote the same source back with
  `--expected-current-sha256`.
- `live-pine-restore.json` - restored from the created backup manifest.
- `live-pine-errors.json` - structured Monaco diagnostics shape.
- `live-pine-safety-summary.json` - compact smoke summary.
- `mcp-tools-list.output.jsonl` / `mcp-tools-list.summary.json` - confirms
  `pine_restore_source` is registered.

Summary:

```json
{
  "source_sha256": "e675eb546b9f96fb79c1b7cd179d8908e7a77c7fcb4ca798124166042d350765",
  "script_name": "EMA Slope Angle + MACD (Michael) v3",
  "script_type": "strategy",
  "pine_version": "6",
  "line_count": 540,
  "char_count": 33307,
  "backup_created": true,
  "restore_verified": true
}
```

No disposable strategy was added to the chart in this smoke. The live write was
limited to setting the same source text back into the editor, then restoring
from the generated backup manifest.

`pine_compile` was not run live in this smoke because the current TradingView UI
can save or add the active script. The implementation now returns structured
diagnostics from Monaco markers after compile; `live-pine-errors.json` records
the same diagnostic response shape without triggering compile/add.

## Tests

```text
go test ./internal/tools/pine ./internal/mcp
go test ./...
go vet ./...
```

All passed on 2026-04-27.

`tools/list` returned 85 tools and included `pine_restore_source`.
