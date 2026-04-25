# TODO — tradingview-mcp-go stabilization

## Status baseline (2026-04-25)

```text
tradingview-mcp-go: parity baseline passed
78/78 tools implemented
CLI implemented
MCP manual implementation accepted
critical blockers: none
```

---

## Phase 1 — Golden tests (MCP protocol layer)

Tests that do NOT require TradingView running. Pure protocol coverage.

- [ ] `initialize` → correct protocolVersion, serverInfo, capabilities
- [ ] `tools/list` → returns exactly 78 tools, each has name/description/inputSchema
- [ ] `tools/call` known tool → `{"content":[{"type":"text","text":"..."}],"isError":false}`
- [ ] `tools/call` unknown tool → `{"content":[...],"isError":true}`
- [ ] `tools/call` with bad JSON args → `{"content":[...],"isError":true}`
- [ ] `ping` → `{}`
- [ ] unknown method → JSON-RPC error -32601
- [ ] parse error (malformed JSON) → JSON-RPC error -32700
- [ ] multiline JSON input (json.Decoder regression test)
- [ ] large response (>64KB) — regression for old Scanner limit

---

## Phase 2 — Smoke tests (TradingView live)

Require: TradingView Desktop running with `--remote-debugging-port=9222`.
Run manually or in a dedicated CI environment with TradingView.

- [ ] CDP connect → `tv status` returns connected=true
- [ ] `chart_get_state` → returns symbol, timeframe, type
- [ ] `quote_get` → returns bid/ask/last, numeric values
- [ ] `data_get_ohlcv` → returns ≥1 bar with time/open/high/low/close/volume
- [ ] `data_get_study_values` → returns study list (may be empty if no indicators)
- [ ] `capture_screenshot` → returns base64 or file path, non-empty
- [ ] `pine_get_source` → returns string (may be empty)
- [ ] `tv_launch` → starts TradingView process, returns success=true

---

## Phase 3 — Windows doctor

`tv doctor` should provide actionable diagnostics on Windows.

- [ ] Check `localhost:9222` reachable — clear message if not
- [ ] List running processes, detect TradingView.exe without `--remote-debugging-port`
- [ ] Probe WindowsApps/MSIX path (`Get-AppxPackage *TradingView*`)
- [ ] Probe `%LOCALAPPDATA%\TradingView\` and `%APPDATA%\TradingView\`
- [ ] Output exact command to restart with CDP flag
- [ ] Detect if port 9222 is used by a different process (Chrome, etc.)
- [ ] Structured JSON output + human-readable hint block

---

## Phase 4 — HTS-ready tools

New tools to support the HTS integration layer.

- [ ] `chart_context_for_llm` — combines chart_get_state + quote_get + top-N study values into one call; returns a single structured object ready for LLM prompt injection
- [ ] `indicator_state` — current value + signal direction (above/below zero, crossing) for a named indicator; reduces LLM need to interpret raw arrays
- [ ] `market_summary` — symbol, timeframe, last bar OHLCV, change%, volume vs avg, current indicators summary; one call for full context
- [ ] `continuous_contract_context` — for futures: nearest expiry, roll date, front/back spread (read from TradingView visible data)

---

## Phase 5 — JSON contracts (HTS integration)

Formal response schemas for tools the HTS layer consumes.
See: [JSON_CONTRACTS.md](JSON_CONTRACTS.md)

Priority tools:

- [ ] `data_get_study_values` — finalize field names, handle empty/null study arrays
- [ ] `data_get_indicator` — clarify `values` array shape, index 0 = current bar
- [ ] `chart_get_state` — confirm `indicators` array structure
- [ ] `quote_get` — confirm all numeric fields present even when bid/ask unavailable
- [ ] `symbol_search` / `symbol_info` — confirm `type`, `exchange`, `description` always present
- [ ] Define which errors are **retryable** (CDP disconnect) vs **permanent** (unknown tool)

---

## Technical debt

- [x] Replace `bufio.Scanner` with `json.Decoder` in mcp/server.go (done 2026-04-25)
- [ ] MCP server: add max message size guard (`io.LimitedReader`, e.g. 16 MB)
- [ ] CDP client: configurable reconnect backoff (currently hardcoded)
- [ ] `go test ./...` in CI (add to release.yml before package step)
- [ ] `.github/workflows/test.yml` — run tests on every push to master
