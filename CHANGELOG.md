# CHANGELOG.md

Формат основан на Keep a Changelog, но адаптирован под процесс портирования Node.js → Go.

Каждый этап портирования должен добавлять датированную запись. Типы записей: **Added**, **Changed**, **Fixed**, **Compatibility**, **Pending**, **Breaking** (Breaking — только с обоснованием).

---

## 2026-04-28 (`Release 1.2 hardening: compatibility probes and equity policy`)

### Added

- `tv_discover` now returns structured `compatibility_probes` while preserving
  the legacy `paths` object.
- Compatibility probes cover TradingView API root, active chart widget, chart
  model, study model data paths, strategy source paths, backtesting API, and
  explicit Strategy Equity plot availability.
- `research/compatibility-probes/` documents the probe contract, equity
  decision, optional history-load workflow, and release evidence.

### Changed

- EN/RU docs, all agents, and all skills now explicitly state that
  `coverage: loaded_chart_bars` is chart-loaded coverage only.
- Derived equity is documented as conditional and not native Strategy Tester
  equity.
- Full native bar-by-bar Strategy Tester equity is explicitly out of scope
  until TradingView exposes a stable report field.
- Optional history loading is documented as best-effort chart range
  expansion/scrolling followed by repeated data calls and
  `loaded_bar_count` / `data_points` comparison.

### Verified

- Live `tv discover` passed and saved output under
  `research/compatibility-probes/`.
- `go test ./...` passed.
- `go vet ./...` passed.
- MCP initialize, `tools/list` count (`85`), and unknown-tool error shape
  passed; artifacts are in `research/compatibility-probes/`.

## 2026-04-27 (`Gap closure: localized Pine compile, docs, agents, skills`)

### Added

- `research/pine-localized-compile/` live smoke evidence for Russian
  TradingView Add-to-chart button handling.
- `research/implementation-gap-audit/` with research/code/docs coverage matrix
  and agents/skills gap summary.
- New skills:
  `study-model-values`, `strategy-backtesting-api`, `strategy-equity-plot`,
  `pine-safe-edit`, `tradingview-limit-handling`, `regression-smoke`, and
  `data-quality`.
- Russian `SKILL.ru.md` variants for all 18 skills.
- Russian agent variants for root, Codex, Cline, Windsurf, Gemini, Cursor, and
  Continue wrappers.
- Russian source prompts:
  `prompts/market-analyst.ru.md`, `prompts/futures-analyst.ru.md`, and
  `prompts/performance-analyst.ru.md`.

### Changed

- `pine_compile` and `pine_smart_compile` now share a localized Add-to-chart
  matcher that recognizes English and Russian labels, including duplicated
  TradingView text such as `Добавить на графикДобавить на график`.
- Active EN/RU docs now distinguish current Go registry count (`85`) from the
  historical Node.js parity baseline (`78`).
- Existing skills now include current MCP source/reliability/status/coverage
  contract notes.
- Agent prompts and all client wrappers were regenerated from updated
  source-of-truth prompts with explicit data-quality rules.
- `PLAN.md`, `TODO_codex.md`, and `prompts.md` were reset for the current
  gap-closure pass.

### Verified

- `go test ./internal/tools/pine` passed.
- Live Russian UI smoke clicked `Добавить на графикДобавить на график`,
  returned `study_added:true`, then removed the disposable strategy and
  restored the original Pine source SHA256.
- `go test ./...` and `go vet ./...` passed.
- MCP initialize, `tools/list` count (`85`), and unknown-tool error shape passed;
  artifacts are in `research/agents-skills-sync/`.

## 2026-04-27 (`Implementation final regression`)

### Verified

- Final regression artifacts were collected in
  `research/implementation-final/`.
- `go test ./...` and `go vet ./...` passed with exit code `0`.
- MCP initialize, `tools/list`, and `tools/call` error shape were verified
  over stdio JSON-RPC. Current registry count is `85` tools.
- Live `data_get_indicator` and `data_get_indicator_history` returned
  RSI/ADX/DI numeric values from `tradingview_study_model` on `RUS:NG1!`.
- A disposable Pine strategy with
  `plot(strategy.equity, "Strategy Equity", display=display.data_window)` was
  added, smoke-tested, and removed.
- Live `data_get_strategy_results`, `data_get_trades`, and `data_get_orders`
  returned `status: ok` from `tradingview_backtesting_api`.
- Live `data_get_equity` returned `400` loaded-bar points from
  `tradingview_strategy_plot` with `coverage: loaded_chart_bars`.
- Pine source was restored from SHA256-verified backup after the smoke test.

### Notes

- TradingView internals remain explicitly marked as unstable internal paths.
- Equity output is not claimed to be full Strategy Tester history unless
  TradingView has loaded the full chart range.
- The final smoke exposed a UI localization gap: the Russian
  `Добавить на график` button had to be clicked directly because the existing
  compile helper recognizes English Add-to-chart labels.

## 2026-04-27 (`Known issues closure`)

### Changed

- `symbol_search` now returns a structured response for empty results:
  `status: no_results`, `reason`, `query`, `count: 0`, and `results: []`.
  The CLI now uses the same response shape instead of printing a bare array.
- `quote_get` now marks unavailable TradingView bid/ask explicitly with
  `bidAskAvailable`, `bidAvailable`, `askAvailable`, `sourceLimitation`, and
  `warning` while preserving numeric `bid`/`ask` sentinel fields for
  compatibility.
- MOEX futures quotes such as `RUS:NG1!` with zero/absent bid/ask are marked
  `sourceLimitation: tradingview_moex_futures_bid_ask_unavailable`.
- `STATUS.md` and `research/known-issues-closure/` now classify Pine,
  strategy, and replay verification as `live_tested`, `partial`, or
  `unverified`.

### Verified

- Before/after evidence is stored in `research/known-issues-closure/`.
- Live `symbol_search` empty query now returns `status: no_results` with a
  reason.
- Live `quote_get` on `RUS:NG1!` returns `bidAskAvailable:false` and a MOEX
  futures source limitation.
- Live `data_get_indicator` / `data_get_indicator_history` on study `Vvzmzg`
  returned RSI/ADX/DI floats from `tradingview_study_model`.
- Live screenshot with a `.png` filename returned
  `screenshots\issue-closure-after-redirect.png`, not `.png.png`.
- `go test ./...` and `go vet ./...` passed.

### Remaining

- Strategy loaded-report extraction remains `partial` in this pass because the
  current chart returned `no_strategy_loaded`; no chart/Pine mutation was
  performed.
- Replay trade workflow remains `unverified`; only `replay_status` was checked.

## 2026-04-27 (`Study limit detection`)

### Added

- `chart_manage_indicator` now accepts optional `allow_remove_any`.
- `tv manage-indicator add NAME [--inputs JSON_ARRAY] [--allow-remove-any]`
  and `tv manage-indicator remove ENTITY_ID` CLI helpers.
- `research/study-limit-detection/` smoke artifacts and README.

### Changed

- Indicator add now verifies that `createStudy` actually produced a new study
  entity. A no-op add is returned as `success:false` instead of silent success.
- TradingView study-limit messages are normalized to
  `status: study_limit_reached` with `currentStudies`, optional `limit`, and
  a user-facing `suggestion`.
- Automatic removal is disabled by default. If `allow_remove_any=true` is
  explicitly passed and a limit is detected, the tool removes the most recent
  existing study, logs the removal to
  `research/study-limit-detection/removals.jsonl`, and retries once.

### Verified

- Unit coverage for English/Russian limit-message parsing, non-limit add
  failures, schema exposure of `allow_remove_any`, and JSONL removal-log
  writing.
- Live smoke on `RUS:NG1!` showed the current session did not hit the 2-study
  limit: `Moving Average` was added as `5C13zY` and explicitly removed during
  cleanup. Artifacts are in `research/study-limit-detection/`.
- `go test ./...` and `go vet ./...` passed.

### Pending

- A live `study_limit_reached` response still needs reproduction on an account
  or chart state where TradingView enforces the study cap.

## 2026-04-27 (`Pine source safety`)

### Added

- `pine_restore_source` MCP tool and `tv pine restore BACKUP_PATH` CLI command.
- `research/pine-source-safety/` live smoke artifacts and README with the safe
  workflow: get source, backup, set source, compile, verify, restore.

### Changed

- `pine_get_source` now returns source metadata:
  `source_sha256`, `hash`, `script_name`, `script_type`, `pine_version`,
  `line_count`, and `char_count`.
- `pine_set_source` now creates a backup before writing to Monaco and returns
  `backup_path`, `backup_source_path`, and `backup_source_sha256`.
- `pine_set_source` accepts optional `expected_current_sha256` as a write guard;
  CLI equivalent is `--expected-current-sha256`.
- `pine_new` and `pine_open` also backup the current editor before replacing
  its content.
- `pine_compile`, `pine_smart_compile`, and `pine_get_errors` now return
  structured diagnostics with `error_count`, `warning_count`, `errors`,
  `warnings`, and `diagnostics`.
- MCP registry count is now 85 tools after adding `pine_restore_source`.

### Verified

- Live `tv pine get` returned script metadata for
  `EMA Slope Angle + MACD (Michael) v3` with SHA256
  `e675eb546b9f96fb79c1b7cd179d8908e7a77c7fcb4ca798124166042d350765`.
- Live `tv pine set-file` wrote the same source using
  `--expected-current-sha256` and created a backup manifest under
  `research/pine-source-safety/session-*/backup.json`.
- Live `tv pine restore` restored from the backup manifest and verified the
  final editor SHA256.
- `go test ./...` and `go vet ./...` passed.

### Pending

- A disposable strategy was not added/compiled in this pass to avoid mutating
  chart studies. The safety workflow was verified by writing the same source
  back into the editor and restoring it from backup.

## 2026-04-27 (`Strategy equity plot extraction`)

### Changed

- `data_get_equity` now first looks for an explicit `Strategy Equity` Pine plot
  in the active strategy source through `metaInfo().plots` and
  `metaInfo().styles`.
- When found, equity is read from
  `model.strategySources()[0].data().fullRangeIterator()` and returned as
  `{index, time, equity}` points with `time` normalized to milliseconds.
- Successful plot extraction is marked with
  `source: tradingview_strategy_plot`,
  `coverage: loaded_chart_bars`, and
  `reliability: reliable_pine_runtime_value_unstable_internal_path`.
- If the plot is absent, `data_get_equity` returns
  `status: needs_equity_plot` plus the suggested Pine line:
  `plot(strategy.equity, "Strategy Equity", display=display.data_window)`.
- Derived fallback metadata is now explicitly marked as
  `source: derived_from_ohlcv_and_trades`,
  `reliableForTradingLogic: false`, with requirements and limitations. Any
  trade-exit fallback points are marked `coverage: trade_exit_points_only`.

### Verified

- Unit coverage checks the equity JS builder for `fullRangeIterator`,
  `Strategy Equity`, `needs_equity_plot`, `tradingview_strategy_plot`,
  `derived_from_ohlcv_and_trades`, and loaded-bar coverage markers.
- Current live chart has no loaded strategy, so the live MCP smoke returned
  `status: no_strategy_loaded`; artifacts are in
  `research/strategy-equity-extraction/`.
- `go test ./...` and `go vet ./...` passed.

### Pending

- Re-run `data_get_equity` with the disposable strategy from
  `research/strategy-equity-full/mcp-test-sma-strategy-equity-plot.pine`.
  This was not done in this implementation pass to avoid mutating the current
  chart/Pine state.

## 2026-04-27 (`Strategy tools use backtesting API`)

### Added

- `data_get_orders` MCP tool for `report.filledOrders` from TradingView's
  backtesting report.
- `research/strategy-backtesting-api/` MCP smoke artifacts for strategy tools.

### Changed

- `data_get_strategy_results`, `data_get_trades`, and `data_get_equity` now use
  `model.strategySources()`, `model.activeStrategySource()`, and
  `await window.TradingViewApi.backtestingStrategyApi()`.
- Strategy extraction no longer detects strategies through
  `dataSources() + performance`, because ordinary studies can expose
  `performance`.
- Strategy tool responses now include explicit statuses:
  `ok`, `no_strategy_loaded`, `strategy_report_unavailable`,
  `strategy_report_shape_unverified`, and
  `tradingview_backtesting_api_unavailable`.
- `data_get_equity` no longer probes fake strategy source data; dedicated
  plot-based equity extraction is documented in the later strategy equity
  entry above.
- MCP registry count is now 84 tools after adding `data_get_orders`.

### Verified

- Current live chart with ordinary indicators only returns
  `success:false/status:no_strategy_loaded` for strategy results, trades,
  orders, and equity.
- `tools/list` includes `data_get_orders`.
- `go test ./...` and `go vet ./...` passed.

### Pending

- Re-run the new MCP tools with a live loaded disposable strategy. Earlier
  research in `research/strategy-live-test/` confirmed the report contains
  `trades`, `filledOrders`, `performance`, `settings`, and `currency`, but this
  implementation smoke did not mutate the current chart/Pine state.

## 2026-04-27 (`Indicator study model extraction`)

### Added

- `data_get_indicator_history` MCP tool for loaded-bar indicator history from
  TradingView's internal study model.
- `tv indicator-history NAME_OR_ENTITY_ID [--count N]` CLI helper for the same
  history extraction path.
- `research/indicator-study-model/` live smoke artifacts for
  `data_get_study_values`, `data_get_indicator`, and
  `data_get_indicator_history`.

### Changed

- `data_get_study_values` and `data_get_indicator` now read numeric plot values
  from `study.data().valueAt(index)` instead of Data Window/UI display strings.
- Plot values are mapped through `metaInfo().plots` and `metaInfo().styles`;
  colorer/alertcondition plots and hidden plots are not exposed as numeric
  study outputs.
- Study-model responses are marked with
  `source: tradingview_study_model`,
  `reliability: reliable_pine_runtime_value_unstable_internal_path`,
  `reliableForTradingLogic: true`, and `coverage: loaded_chart_bars`.
- MCP registry count is now 83 tools after adding `data_get_indicator_history`.

### Verified

- Live `RUS:NG1!` 1D RSI/ADX/DI values from study `Vvzmzg` returned raw floats:
  RSI `31.141247635850746`, ADX `26.293855814935608`,
  DI+ `15.545741865281482`.
- History smoke returned 5 loaded bars via `fullRangeIterator()`.
- `go test ./...` and `go vet ./...` passed.

## 2026-04-27 (`Strategy live extraction test`)

### Added

- `research/strategy-live-test/` — live TradingView Desktop research artifacts
  for adding `MCP Test SMA Strategy` and reading
  `window.TradingViewApi.backtestingStrategyApi()` without CSV export.
- CLI-only `tv ui eval-await` for CDP `Runtime.evaluate` with
  `awaitPromise=true`; existing MCP `ui_evaluate` behavior is unchanged.
- CLI-only `tv pine set-file PATH` for restoring large Pine sources without
  Windows command-line length issues; existing MCP Pine tools are unchanged.

### Verified

- A simple Pine strategy was added after freeing the Basic subscription
  indicator limit.
- `backtestingStrategyApi()` saw the active strategy and returned non-null
  `activeStrategyReportData`.
- `activeStrategyReportData` contained `trades`, `filledOrders`,
  `performance`, `buyHold`, and `runupDrawdownPeriods`.
- Normalized extraction prototype produced 43 trades, 86 filled orders,
  summary metrics, and a close-to-close equity reconstruction.

### Notes

- Full bar-by-bar strategy equity was not found in `activeStrategyReportData`;
  current equity reconstruction uses `trades[].cumulativeProfit` plus initial
  capital at trade exit times.
- `Help system for trade` was removed to free the Basic limit slot. The test
  strategy was removed and Pine source was restored exactly, but the removed
  custom indicator could not be re-added through `createStudy`; `Volume` was
  re-added instead.

## 2026-04-27 (`Strategy internals research`)

### Added

- `results/strategy-internals-research-2026-04-27.md` - live CDP research of
  TradingView strategy internals, including exact chart model paths,
  `model.strategySources()`, `model.activeStrategySource()`, and
  `TradingViewApi.backtestingStrategyApi()` watched values.

### Confirmed

- The current chart has no loaded strategy:
  `strategySources().length == 0`, active strategy is `null`, backtesting
  report data is `null`, and `isStrategyEmpty` is `true`.
- Ordinary studies expose a `performance` property, so the existing
  `dataSources()` strategy detector is too broad and can select non-strategy
  studies.

### Pending

- Verify the concrete `activeStrategyReportData` shape with a known loaded
  disposable strategy before extracting trades, orders, PnL, and equity as
  reliable data.
- Update strategy tools to return explicit `no_strategy_loaded` /
  `strategy_report_unavailable` statuses instead of empty successful results.

## 2026-04-27 (`Final HTS audit report`)

### Added

- `docs/dev/FINAL_AUDIT_REPORT.md` — final HTS readiness audit with executive
  summary, readiness matrix, blocking issues, non-blocking issues, next actions,
  safe-to-use tools, tools requiring work, and decision.

### Decision

- Final research-stage decision: `GO WITH LIMITATIONS`.
- `tradingview-mcp-go` is suitable as a TradingView chart/context/visual/Pine
  collection layer, while HTS Go must calculate critical features and Tinkoff
  must provide execution-grade instrument identity, orderbook, expiration,
  margin, and trading status.

### Changed

- `STATUS.md` now reflects current evidence: `symbol_search` and screenshot
  filename are fixed, and indicator-value risk is classified as Data
  Window/UI text parsing until source flags/history mode are implemented.

## 2026-04-27 (`Instrument Resolver Contract`)

### Added

- `docs/dev/INSTRUMENT_RESOLVER_CONTRACT.md` — draft documentation-only
  contract for resolving TradingView analysis symbols to concrete Tinkoff
  execution instruments.
- Proposed external `InstrumentResolver` and `TinkoffMarketData` interfaces,
  Go structs, SQLite schema, CLI commands, MCP tools, warning codes, and
  fail-closed fallback behavior.
- Explicit rule that TradingView continuous futures are analysis-only, while
  Tinkoff concrete futures provide execution identity, orderbook bid/ask/spread,
  expiration, margin, and trading status.

### Boundary

- Resolver storage, Tinkoff API calls, orderbook, margin/status retrieval, and
  execution readiness stay in the external HTS MCP layer, not in
  `tradingview-mcp-go`.

## 2026-04-27 (`LLM Market Context Contract`)

### Added

- `docs/dev/LLM_MARKET_CONTEXT_CONTRACT.md` — draft documentation-only
  contract for compact LLM payloads that avoid raw candles by default.
- JSON Schema for `instrument_summary`, `market_scan_summary`, `diff_summary`,
  `trade_review_summary`, shared warning/quality objects, and strict
  `hts.llm_response.v1` output.
- Proposed external Go structs and prompt templates for DeepSeek, Qwen, Kimi,
  Opus, and ChatGPT.
- Forbidden practices list covering raw candles, canvas coordinates, DOM text
  numerics, TradingView bid/ask sentinels, continuous futures as execution
  symbols, unverified Pine output, and free-form LLM responses.

### Boundary

- The document defines the external HTS MCP -> LLM interface only; it does not
  implement HTS orchestration, broker integration, trading execution, or risk
  sizing inside `tradingview-mcp-go`.

## 2026-04-27 (`HTS Market Summary Contract`)

### Added

- `docs/dev/HTS_MARKET_SUMMARY_CONTRACT.md` — draft documentation-only
  contract for TradingView MCP -> external HTS MCP -> LLM market summaries.
- The contract separates `tradingview_direct`, `hts_go_derived`,
  `pine_hts_json`, `tinkoff_marketdata`, and `unreliable` source classes.
- Proposed external Go structs, a JSON example, `data_quality`,
  `source_trace`, warnings, and a list of fields that must not be sent to an
  LLM as factual truth.

### Boundary

- Tinkoff integration, execution symbol resolution, risk sizing, and trading
  decisions remain outside `tradingview-mcp-go`; this repository only documents
  the contract boundary and TradingView-sourced inputs.

## 2026-04-27 (`Screenshot filename normalization`)

### Fixed

- `capture_screenshot` now appends `.png` only when the requested filename has
  no extension.
- Existing `.png` / `.PNG` filenames are preserved, so `foo.png` no longer
  becomes `foo.png.png`.
- Non-PNG extensions now return an explicit error instead of being silently
  rewritten, because the tool always writes PNG content and changing a caller's
  requested extension can hide mistakes or unexpected output paths.

### Added

- Unit tests for screenshot filename normalization, Windows/path-like names,
  default generated names, and non-PNG extension rejection.

## 2026-04-26 (`MOEX bid/ask source research`)

### Added

- `results/bid-ask-moex-futures-research-2026-04-26.md` — live research
  report for why `quote_get` returns `bid:0` / `ask:0` on MOEX natural gas
  futures.

### Findings

- `RUS:NG1!` and front contract `RUS:NGJ2026` do not expose `bid` / `ask`
  fields in TradingView `mainSeries().quotes()` in the tested session;
  quote state reports last-trade/OHLCV fields only.
- TradingView can expose BBO on other symbols: `NYMEX:NG1!` and
  `BINANCE:BTCUSDT` had internal `bid`, `ask`, `bid_size`, and `ask_size`.
- `quote_get` should read internal `mainSeries().quotes().bid/ask` when
  present and add explicit availability/source fields. For MOEX futures with
  absent keys, keep numeric fields as compatibility sentinels and mark
  bid/ask unavailable.
- HTS should use Tinkoff Invest API orderbook for executable MOEX futures
  bid/ask/spread, outside this TradingView MCP repository.

## 2026-04-26 (`Strategy filter hypothesis workflow`)

### Added

- `results/strategy-filter-hypothesis-workflow-2026-04-26.md` — architecture
  for LLM-proposed additional strategy filters as testable hypotheses, not
  direct strategy edits.

### Findings

- LLM should only propose formalized hypotheses for filters such as ADX,
  ATR volatility regime, EMA/KAMA trend, RSI zones, volume confirmation,
  time/session, distance-to-level, no-trade zones, and trend phase filters.
- Every hypothesis must pass formalization, baseline backtest comparison,
  out-of-sample/forward testing, and manual acceptance.
- Existing MCP/CLI tools can collect baseline context, OHLCV summaries,
  screenshots, Pine source, and prepared Pine outputs, but normalized strategy
  trades/equity and CLI wrappers for strategy data/input mutation are still
  missing.

## 2026-04-26 (`Strategy diagnosis LLM data research`)

### Added

- `results/strategy-diagnosis-llm-data-research-2026-04-26.md` — research
  report defining the minimum structured data an LLM needs to diagnose weak
  strategy performance through MCP/HTS.

### Findings

- Current MCP is insufficient for reliable strategy diagnosis by itself:
  source/inputs/OHLCV/chart context are partially available, but normalized
  trades, equity curve, drawdown curve, robust Strategy Tester metrics, and
  per-trade regime annotations are not stable.
- `data_get_strategy_results`, `data_get_trades`, and `data_get_equity` need
  explicit unavailable/no-strategy statuses before their output can be consumed
  as evidence by an LLM.
- Recommended path: MCP collects TradingView data, HTS prepares
  `strategy_review_summary`, and the LLM receives only compact summaries,
  failure clusters, top hypotheses, and robustness checks.

### References

- TradingView Strategy Report CSV export and Strategy Report metrics were used
  as the preferred external source for trades and performance metrics.
- Pine `strategy.*` trade/equity outputs remain the preferred prepared-script
  path when source can be edited legally.

## 2026-04-26 (`Indicator input control stabilization`)

### Added

- `results/indicator-input-control-research-2026-04-26.md` — research and
  implementation report for indicator/strategy input control through MCP.
- `tests/smoke`: `TestIndicatorSetInputsVolumeLength` mutates built-in
  `Volume.length`, verifies the value through `data_get_indicator`, and
  restores the original value.
- `Makefile`: `smoke-indicator-input` target for the focused live smoke test.

### Changed

- `indicator_set_inputs` now verifies requested input IDs and returns
  `changed_inputs`, `unchanged_inputs`, `missing_input_ids`,
  `failed_input_ids`, `changes`, `applied_count`, and `changed_count`.
- `indicator_set_inputs` now returns `success:false` when none of the requested
  input IDs exist, instead of reporting fake success.
- `indicator_set_inputs` masks very large string values in before/after change
  metadata to avoid echoing internal Pine payloads.
- `indicator_toggle_visibility` now returns `before_visible`, `visible`,
  `changed`, `entity_id`, and `source` while preserving `entityId`.

### Verified

- Built-in `Volume.length` changed live from `20` to `21` and was restored to
  `20`.
- Custom Pine `Помошник RSI - True ADX` boolean input `in_0` changed live from
  `true` to `false` and was restored to `true`.
- `indicator_toggle_visibility` changed `Volume` visibility off and back on.
- Unknown input id now returns an explicit failure with `missing_input_ids`.

## 2026-04-26 (`Indicator and strategy history research`)

### Added

- `results/indicator-strategy-history-research-2026-04-26.md` — research
  report on historical calculated indicator values, Strategy Tester history,
  trades, equity curve, Pine table/label history, UI export, and
  network/WebSocket feasibility.

### Verified

- Built-in `Volume` history is available through the internal chart study model
  via `study.data().valueAt(index)`.
- Custom Pine `Помошник RSI - True ADX` history is available through the same
  internal model with raw float values for DI+/DI-/ADX/RSI.
- Current `data_get_pine_tables` can read current table values from the custom
  RSI/ADX study.
- Current `data_get_strategy_results`, `data_get_trades`, and
  `data_get_equity` do not provide usable strategy history in the current chart
  because no strategy is loaded.

### Findings

- Current public indicator tools expose only Data Window snapshots, not history.
- A new `data_get_indicator_history` helper/tool should read
  `study.data().valueAt(index)` and map rows through `metaInfo().plots`.
- Strategy tools need explicit `no_strategy_loaded` / unavailable statuses and
  should not treat regular indicator `performance` properties as a strategy.
- For HTS, prefer internal TradingView study history for applied indicators,
  prepared Pine `HTS_JSON` for custom strategy/trade state, and Go-side
  recalculation from OHLCV for deterministic built-ins.

## 2026-04-26 (`TradingView account limits research`)

### Added

- `results/tradingview-account-limits-research-2026-04-26.md` — research
  report on account/plan detection, indicator-per-chart limits, empirical
  limit probing, CDP capture of TradingView limit UI, and recommended MCP error
  contract.

### Verified

- Current TradingView session classified as Basic/free through page state and
  profile API (`window.pro.isPro() == false`, empty `pro_plan`).
- Official pricing mapping checked on 2026-04-26: Basic=2, Essential=5,
  Plus=10, Premium=25, Ultimate=50 indicators per chart.
- Empirical add probe `chart_manage_indicator add Moving Average` did not add a
  new study and left chart study count unchanged.
- TradingView showed a localized limit dialog that CDP can read from the DOM:
  the current subscription allows 2 applied indicators.

### Findings

- Current `chart_manage_indicator` add behavior is unsafe: it can return
  `success:true` with an empty `entityId` when TradingView rejected the add.
- Limit detection should combine plan inference, official plan mapping,
  before/after study id comparison, and DOM/dialog text capture.
- If plan is unknown, MCP/HTS should use a conservative fallback and avoid
  adding more than 2 indicators automatically.

## 2026-04-26 (`Pine source access research`)

### Added

- `results/pine-source-access-research-2026-04-26.md` — research report on
  Pine source availability for current editor scripts, saved user scripts,
  open-source publications, protected/invite-only scripts, built-ins, and
  inputs/outputs without source.

### Verified

- MCP `pine_get_source` live returned `33307` chars / `540` lines from the
  currently open Pine Editor buffer.
- `pine_list_scripts` live returned `43` saved scripts through TradingView
  `pine-facade` page-context credentials.
- Existing `pine_open` is source-capable for saved user scripts, but mutates the
  editor buffer after fetching source.
- `data_get_indicator` can expose inputs and plots without source.

### Findings

- MCP can return Pine source only when TradingView already exposes that source
  to the current account/session.
- Protected and invite-only source must remain unavailable for non-author users;
  MCP should return explicit `source_available:false` and may expose permitted
  inputs/outputs only.
- Many Pine-based built-ins expose source through TradingView's normal Pine
  Editor/source-code UI; non-Pine built-ins do not expose Pine source.
- LLM-facing source transfer needs explicit source origin/access metadata and
  optional secret redaction.

## 2026-04-26 (`indicator value source research`)

### Added

- `results/indicator-values-source-research-2026-04-26.md` — live research
  report covering Data Window, DOM, internal chart model, Pine outputs, plot
  mapping, strategy state, network/WebSocket feasibility, canvas fallback, and
  proposed `HTS_JSON` format.

### Findings

- Correct calculated indicator values and history are available in the current
  TradingView Desktop page through the internal chart widget model:
  `study.data().valueAt(index)`.
- The current wrong values are not canvas/Y coordinates in the tested setup.
  They are formatted Data Window / legend strings parsed incorrectly
  (`31,35` -> `3135`, `17,27 K` -> `1727`).
- Observed plot mapping contract: `row[0] = time` and
  `row[plot_array_index + 1] = value for metaInfo().plots[plot_array_index]`.
  `colorer`, `alertcondition`, fill, and hidden plots still occupy row slots;
  filtering must happen after mapping.
- Data Window and DOM legend are useful display fallbacks, but are rounded,
  localized, and sometimes scaled.
- Pine tables/labels are the best deterministic source for custom scripts when
  they emit a strict machine-readable `HTS_JSON:` payload.
- Network/WebSocket capture is not recommended as the primary value source;
  page JS does not expose historical WS payloads.
- Screenshot/canvas reverse mapping is unnecessary and remains unsafe for
  numerical decisions.

## 2026-04-26 (`Pine / strategy / replay live audit`)

### Added

- `results/live-pine-strategy-replay-2026-04-26/` — separate live-test result
  files for all requested Pine, strategy, and replay tools.
- `results/live-pine-strategy-replay-2026-04-26/SUMMARY.md` — readiness table,
  blocking issues, safe-to-use list, and tools not safe for HTS without changes.

### Verified

- `pine_get_source` — live CDP/Monaco read OK (`33307` chars, `540` lines).
- `pine_set_source`, `pine_new`, `pine_open` — live editor mutation works.
- `pine_list_scripts` — live pine-facade list OK (`43` scripts).
- `pine_analyze` — offline static analyzer works; explicitly not TradingView API.
- `pine_check` — TradingView pine-facade `translate_light` API compiled a sample.
- Replay tools — `replay_start`, `replay_step`, `replay_autoplay`,
  `replay_status`, and `replay_stop` worked against `RUS:NG1!` 1D.

### Findings

- `pine_compile` is unsafe in the tested UI state: it clicked `Pine Save`,
  saved a test source over the active saved script, and did not add a study to
  the chart. The original script was restored from pine-facade version `25.0`
  and saved back as version `27.0`.
- `pine_smart_compile` is partial: it returned `has_errors:false` and
  `study_added:false`, with `button_clicked:"Pine Save"`.
- `pine_get_console` is a noisy DOM scrape, not a structured compile log.
- `data_get_strategy_results` and `data_get_equity` returned `success:true`
  with empty data when no strategy was loaded; they need explicit
  `no_strategy_loaded` / unavailable status.
- `data_get_trades` included `error:"No strategy found on chart."` but still
  used `success:true`.
- `replay_trade` returned `success:true`, but `position` and `realized_pnl`
  stayed `null`, so trade execution/state confirmation is weak.

## 2026-04-26 (`symbol_search` stabilization)

### Fixed

- `internal/tools/chart/symbol.go` — fixed `symbol_search` returning `[]`.
  Root cause: the request sent `type=` / `exchange=` and `search_type=undefined`
  to TradingView's v3 symbol-search API. The API now rejects any `type` parameter
  with `forbidden_set_type_with_search_type_api`.
- `symbol_search` now omits unsupported params, sends `exchange` only when set,
  and applies `type`/`exchange` filtering client-side after parsing results.
- HTTP non-2xx responses now return an explicit error with status code and a
  short body snippet instead of silently parsing into an empty result set.
- Search result parsing now supports both current v3 object responses
  (`{"symbols":[...]}`) and legacy array responses.
- MCP `symbol_search` response includes `source: "tradingview_symbol_search_api"`.

### Added

- Unit coverage for symbol-search URL construction, v3/legacy response parsing,
  client-side filtering, and `<em>` stripping.
- `tests/smoke`: `TestSymbolSearch` checks that `NG` returns non-empty results
  with required fields when live smoke tests are run.
- `Makefile`: `smoke-symbol-search` target runs `go run ./cmd/tv symbol-search NG`.

### Verified

- `go test ./internal/tools/chart` — PASS.
- CLI checks:
  - `go run ./cmd/tv symbol-search NG`
  - `go run ./cmd/tv symbol-search NG --exchange RUS`
  - `go run ./cmd/tv symbol-search NG --type futures`
  - `go run ./cmd/tv symbol-search BTCUSD`
  - `go run ./cmd/tv symbol-search SBER`
- Page-context check through `tv ui eval`: TradingView page can reach the same
  endpoint (`status=200`, 50 results for `NG`), so a CDP fetch fallback is viable
  if direct host HTTP is blocked later.

---

## 2026-04-26 (Phase 5)

### Added

- `internal/mcp/errors.go` — new file: `IsRetryable(err) bool` and `ClassifyError(err) ErrorKind`.
  Retryable: `CDP`, `connect`, `no TradingView`, `timeout`, `websocket`.
  Permanent: `unknown tool`, `unmarshal`, `invalid`, `is required`.
  `ErrorKind` constants: `ErrKindCDPDisconnect`, `ErrKindTabNotFound`, `ErrKindJSTimeout`,
  `ErrKindUnknownTool`, `ErrKindBadArgs`, `ErrKindUnknown`.
- `internal/mcp/errors_test.go` — tests for `IsRetryable` (11 cases) and `ClassifyError` (9 cases).

### Added — Skills (Phase 5)

- `skills/json-contracts/SKILL.md` — contract verification guide: field presence, type, and
  sentinel-value checks for all 6 priority tools; array safety patterns; multi-output indicator access.
- `skills/error-handling/SKILL.md` — error classification and recovery guide: retryable vs
  permanent decision tree; retry pattern with exponential wait; CDP disconnect diagnosis flow;
  Windows MSIX note; common error message table.

### Changed — `data_get_study_values`

- JS expression updated: each study entry now includes `entity_id` (string-coerced from
  `s.id` observable), `plot_count` (integer), and `plots` array.
- Each `plots` entry: `{name: string, current: float64|null, values: [float64]}`.
  `values[0]` === `current` (current bar); `values` is `[]` when `current` is null.
- Go type `StudyResult` and `StudyPlot` exported for use by the HTS package.
- `studies` always `[]` (never null) — nil slice guard added.
- `internal/tools/data/data_contracts_test.go` — new tests: JSON shape, nil current, empty
  array guard, `quote_get` sentinel fields, `symbol_info` sentinel fields.

### Changed — `data_get_indicator`

- JS expression updated: `name` now read from `meta.description`/`shortDescription`.
- `inputs` converted from raw `getInputValues()` array to key→value `{}` object.
  Oversized strings still omitted (>500 chars) or truncated (>200 chars for `text` input).
- Added `plots` array from `dataWindowView` of the matching source (same format as
  `data_get_study_values`); `plots` always `[]` if no visible outputs.
- Go-side sentinel: `inputs`, `plots`, `name` always present in response even if JS partial.

### Changed — `chart_get_state`

- JS expression updated: now returns `exchange` (prefix of `symbol` before `:`),
  `ticker` (suffix after `:`), `pane_count` (from `chart._chartWidget.model().panes().length`).
- Go response map: added `exchange`, `ticker`, `pane_count`; added `indicators` as
  canonical alias for `studies` (both present; `studies` kept for backward compat).
- `studies` nil guard — always `[]`.
- `internal/tools/chart/chart_contracts_test.go` — new tests: exchange/ticker parsing,
  contract field presence, `StudyInfo` JSON shape, `symbol_info` sentinels,
  `SymbolSearchResult` empty-field guarantee.

### Changed — `quote_get`

- JS expression updated: calculates `change` = `last[4] - prev[4]` and `change_pct` =
  `(last[4] - prev[4]) / prev[4] * 100` from `bars.valueAt(lastIdx - 1)`.
- `bid` and `ask` DOM scraping now uses `isNaN` guard; both initialised to `0`.
- Go-side sentinel: `bid`, `ask`, `change`, `change_pct` always set to `0.0` if nil.

### Changed — `symbol_info`

- After JSON unmarshal, sentinel loop ensures `symbol`, `exchange`, `description`, `type`
  are always present (empty string default).

### Changed — `hts.go` (internal)

- `ChartContextForLLM`: reads `sv["studies"].([]data.StudyResult)` (typed slice);
  builds `context_text` from `sr.Plots[0].Current` instead of `values` map.
- `IndicatorState`: reads typed `[]data.StudyResult`; `primary_value` / `direction` /
  `signal` from `matched.Plots[0].Current`; response field `plots` replaces `values`.
- `MarketSummary`: reads typed `[]data.StudyResult` for indicators.

### Changed — Documentation

- `docs/en/cli.md` — new section `## JSON Contracts and Error Handling (Phase 5)`:
  contract tables for all 6 tools; retryable vs permanent error classification table.
- `docs/ru/cli.md` — same section in Russian (`## JSON-контракты и обработка ошибок`).
- `prompts/market-analyst.md` — added `## Phase 5 — JSON contract awareness` section.
- `agents/market-analyst.md` + all 6 client variants — appended Phase 5 contract notes.
- `agents/futures-analyst.md` + all 6 client variants — appended Phase 5 contract notes.
- `agents/performance-analyst.md` + all 6 client variants — appended Phase 5 contract notes.
- `agents/README.md` — added `## Skills` table with all 11 skills including 2 new Phase 5 skills.

### Verified

- `go test ./...` — all packages pass (including new tests in `data`, `chart`, `mcp`, `hts`).
- `go build ./...` — clean.

---

## 2026-04-26 (Phase 4)

### Added

- `internal/tools/hts`: new package with 4 HTS-ready composite tools:
  - `chart_context_for_llm` — aggregates `chart_get_state` + `quote_get` +
    top-N `data_get_study_values` into one call; returns `symbol`, `timeframe`,
    `chart_type`, `price` object, `indicators` array, `indicator_count`, and
    `context_text` (compact pipe-separated string for LLM prompt injection).
    `top_n` argument (default 5); study values are best-effort (absent indicators
    do not fail the call).
  - `indicator_state` — finds a study by partial, case-insensitive name match
    against live `data_get_study_values`; returns `matched_name`, `values`,
    `primary_value`, `primary_key`, `direction` (`above_zero`/`below_zero`/`at_zero`),
    `signal` (`bullish`/`bearish`/`neutral`/`overbought`/`oversold`), `near_zero`.
    Signal rules: RSI/Stochastic/Relative Strength Index ≥70=overbought ≤30=oversold;
    CCI ≥100/≤−100; all others positive=bullish/negative=bearish.
  - `market_summary` — one-call full context: symbol, timeframe, chart type,
    last bar OHLCV (`data_get_ohlcv` 21 bars), `change` (close − prev close),
    `change_pct`, `volume_vs_avg` (last ÷ 20-bar prior average), all active
    indicators (best-effort).
  - `continuous_contract_context` — detects `!`-suffixed continuous contract symbols
    (e.g. `NG1!`, `ES1!`, `CL2!`); parses `base_symbol` and `roll_number`; enriches
    with `description`, `exchange`, `type`, `currency_code`, `root_description`
    from `chart.SymbolInfo()` (best-effort). Includes explanatory `note` about
    JS API limitation re: expiry/roll dates.
- `cmd/tvmcp/main.go` — registers `hts.RegisterTools(reg)` (total **82 MCP tools**).
- `cmd/tv/main.go` — CLI commands: `tv context [--top-n N]`, `tv indicator NAME`,
  `tv market`, `tv futures-context`.
- `docs/en/cli.md` — HTS section with argument tables, response field tables,
  signal classification rules, and CLI examples (under `## HTS-ready composite tools`).
- `docs/ru/cli.md` — same section in Russian (under `## HTS-инструменты для LLM`).

### Added — Skills (Phase 4)

- `skills/llm-context/SKILL.md` — LLM Context Builder: one-call chart snapshot via
  `chart_context_for_llm`; validates completeness, embeds `context_text` into prompt chains,
  covers refresh-after-symbol-change pattern.
- `skills/indicator-scan/SKILL.md` — Indicator Signal Scanner: reads all active indicator
  signals by name via `indicator_state`; builds signal table with confluence assessment
  (bullish/bearish/overbought/oversold/mixed); explains `near_zero` as crossover alert.
- `skills/market-brief/SKILL.md` — Market Brief: structured briefing via `market_summary`;
  covers price action, volume classification (`volume_vs_avg`), indicator snapshot,
  multi-symbol iteration pattern, and standard brief format.
- `skills/futures-roll/SKILL.md` — Futures Contract Context: detects continuous contracts via
  `continuous_contract_context`; interprets `base_symbol`/`roll_number`; uses volume drop
  as roll-period proxy; covers front/back spread via dual-pane comparison; includes
  typical roll window notes for energy, equity index, and metals.

### Added — Agents (Phase 4)

Two new agent definitions (`market-analyst`, `futures-analyst`) deployed across all 7 client platforms:

- `agents/market-analyst.md` — Claude Code (Claude Agents SDK, YAML frontmatter)
- `agents/futures-analyst.md` — Claude Code (Claude Agents SDK, YAML frontmatter)
- `agents/cursor/market-analyst.mdc` — Cursor Rules (`.mdc` with YAML frontmatter)
- `agents/cursor/futures-analyst.mdc` — Cursor Rules
- `agents/cline/market-analyst.md` — Cline (`.clinerules/`, plain markdown)
- `agents/cline/futures-analyst.md` — Cline
- `agents/windsurf/market-analyst.md` — Windsurf (`.windsurfrules`, plain markdown)
- `agents/windsurf/futures-analyst.md` — Windsurf
- `agents/continue/market-analyst.prompt` — Continue (`.continue/prompts/`, `name:/description:/---` format)
- `agents/continue/futures-analyst.prompt` — Continue
- `agents/codex/market-analyst.md` — OpenAI Codex CLI (`# name\n\n` + body)
- `agents/codex/futures-analyst.md` — OpenAI Codex CLI
- `agents/gemini/market-analyst.md` — Gemini CLI (`# name\n\n` + body)
- `agents/gemini/futures-analyst.md` — Gemini CLI
- `prompts/market-analyst.md` — canonical source of truth for market-analyst body
- `prompts/futures-analyst.md` — canonical source of truth for futures-analyst body

### Changed — Agents (Phase 4)

- `performance-analyst` Data Gathering step 1 updated in all 7 client formats:
  `chart_get_state` (Phase 3) → `chart_context_for_llm` with `top_n: 3` (Phase 4).
  Step 5 now specifies `region: "chart"` and `region: "strategy_tester"` explicitly.
  Affected files: `agents/performance-analyst.md`, `agents/cursor/performance-analyst.mdc`,
  `agents/cline/performance-analyst.md`, `agents/windsurf/performance-analyst.md`,
  `agents/continue/performance-analyst.prompt`, `agents/codex/performance-analyst.md`,
  `agents/gemini/performance-analyst.md`.
- `agents/README.md` — expanded from 1 agent (performance-analyst) to all 3 agents;
  added agent comparison table; reorganised install instructions by client (all 3 agents
  per section); fixed MD031/MD040 markdown lint warnings.

### Changed

- `internal/mcp/server_test.go` — `TestToolsListExact78` renamed to
  `TestToolsListExact82`; 4 HTS tool names added to `newFullRegistry()`.

### Verified

- `go build ./...` — успешно.
- `go test ./...` — **87/87 тестов PASS** (78 предыдущих + 9 новых в `tools/hts`):
  `TestRegisterHTSToolNames`, `TestParseFirstNumericFloat`,
  `TestParseFirstNumericString`, `TestParseFirstNumericStringWithComma`,
  `TestParseFirstNumericNonNumericString`, `TestParseFirstNumericEmpty`,
  `TestValueDirection`, `TestStudySignalRSI`, `TestStudySignalCCI`,
  `TestStudySignalGeneric`, `TestStrVal`, `TestNumVal`,
  `TestContinuousContractSymbolParsing`, `TestContinuousContractWithExchangePrefix`,
  `TestRound2`, `TestIndicatorStateMissingName`.

---

## 2026-04-26 (Phase 3)

### Added

- `internal/tools/doctor`: new package implementing `Run() *Report` with full
  Windows-aware diagnostics:
  - `PortCheck` — probes `localhost:9222/json/list`; reports reachability,
    whether the response is valid CDP, and the name of any process that owns
    the port via `netstat -ano` + `tasklist`.
  - `ProcessCheck` — detects `TradingView.exe` via `tasklist`; fetches its
    full command line with `Get-CimInstance Win32_Process` to check for
    `--remote-debugging-port`.
  - `InstallCheck` — probes `%LOCALAPPDATA%\TradingView\`,
    `%APPDATA%\TradingView\`, standard install dirs, and MSIX via
    `Get-AppxPackage *TradingView*` (broad wildcard per spec).
  - `LaunchCmd` — exact shell command to restart TradingView with CDP flag.
  - `Hints` — ordered, actionable English strings covering every failure mode
    (not running, wrong port owner, missing install, CDP-less process, etc.).
- `cmd/tv doctor` handler replaced with `doctor.Run()` call; imports cleaned
  up (`cdp` and `discovery` removed from `cmd/tv/main.go`).

---

## 2026-04-26

### Added

- Phase 2 smoke test suite in `tests/smoke/smoke_test.go`:
  `TestCDPConnect`, `TestChartGetState`, `TestQuoteGet`, `TestDataGetOHLCV`,
  `TestDataGetStudyValues`, `TestCaptureScreenshot`, `TestPineGetSource`,
  `TestHealthCheckShape`. All CDP-dependent tests skip gracefully when
  TradingView is not running with `--remote-debugging-port=9222`.

### Fixed

- `internal/discovery`: `findWindowsStore()` now returns `*Result` with
  `IsMSIX`, `MSIXFamilyName`, and `MSIXAppID` fields instead of a plain
  `string`, enabling the launcher to choose the correct activation path.
- `internal/launcher`: MSIX installs on Windows now launch via
  `explorer.exe shell:AppsFolder\<AUMID>` (correct MSIX activation) and
  write `electron-flags.conf` to the app userData directory before launch.
  Direct-installer builds continue to use `--remote-debugging-port=<port>`
  via `exec.Command`.

### Notes

- Microsoft Store (MSIX) version of TradingView Desktop rejects
  `--remote-debugging-port` at the JS argument-parser level and exits.
  CDP is only available when TradingView relaunches itself via
  `app.relaunch()` (e.g. after an auto-update), which preserves the flag
  from the originating process. Use the direct-installer build from
  tradingview.com for reliable CDP access. Phase 3 (tv doctor) will
  detect this and show a clear remediation hint.

---

## 2026-04-25

### Added

- Phase 1 golden tests: 12 protocol-level tests in `internal/mcp/server_test.go`, no TradingView required.
  - `TestInitialize`, `TestToolsListExact78`, `TestToolsCallKnown`, `TestToolsCallUnknown`, `TestToolsCallBadArgs`, `TestPing`, `TestUnknownMethod`, `TestParseError`, `TestNotificationNoResponse`, `TestMultilineJSON`, `TestLargeResponse`, `TestSequentialRequests`.
- `newFullRegistry()` helper with all 78 canonical tool names as stubs — used by `TestToolsListExact78`.
- `.github/workflows/test.yml` — runs `go test ./...` on every push and on pull requests.

### Fixed

- `internal/mcp/server.go`: replaced `json.Decoder` with `bufio.Reader.ReadBytes('\n')` + `json.Unmarshal` per-line. The `json.Decoder` approach caused `TestParseError` to hang (30s timeout) due to state corruption after a failed decode.
- Added 16 MB per-message size guard: lines exceeding `maxMessageBytes` are rejected with `-32700` instead of being unmarshalled.

### Changed

- `TestMultilineJSON` semantics: multiline JSON is a protocol violation in NDJSON. Test now asserts each partial line produces a `-32700` parse error.

---

## 2026-04-24

### Added

- Завершена инвентаризация исходного Node.js репозитория `tradingview-mcp` (P0, P0S).
- Установлено точное число MCP tools: **78**, распределённых по 15 группам.
- Установлено точное число CLI-команд: **83**, распределённых по 16 групп.
- Выявлены 8 CDP методов: `Runtime.evaluate`, `Runtime.enable`, `Page.enable`, `DOM.enable`, `Page.captureScreenshot`, `Input.dispatchKeyEvent`, `Input.insertText`, `Input.dispatchMouseEvent`.
- Задокументированы 6 scripts: `pine_pull.js`, `pine_push.js`, 4 платформенных launch scripts.
- Задокументированы 5 skills/workflows: chart-analysis, multi-symbol-scan, pine-develop, replay-practice, strategy-report.
- Заполнен `COMPATIBILITY_MATRIX.md`: полная таблица tool-by-tool, CLI команды, CDP вызовы, scripts, skills.
- Создан `PORTING_NOTES.md` с архитектурными деталями, паттернами Node.js-кода и примечаниями для Go-порта.
- Обновлён `TODO.md`: исправлены имена tools в P8 (`draw_remove_one`, `draw_get_properties`), добавлены пропущенные tools в P4, P5, P7, P12.
- Добавлено требование поддержки Windows Microsoft Store / WindowsApps установки TradingView Desktop.
- Добавлено требование портировать skills, scripts и utilities, а не только MCP tools.
- Добавлен регламент оркестрации портирования в лимитах Claude Pro.
- Добавлены отдельные документы `WINDOWS_TRADINGVIEW_DISCOVERY.md`, `ORCHESTRATION_AND_LIMITS.md`, `PROMPTS.md`.
- Подготовлен план 1:1 переноса `tradingview-mcp` с Node.js на Go.
- Зафиксирована целевая архитектура Go-проекта.

### Compatibility

- MCP server: `src/server.js` экспортирует 78 tools через `@modelcontextprotocol/sdk`.
- CLI binary: `tv` → `src/cli/index.js`, router → `src/cli/router.js`.
- CDP port default: `localhost:9222`.
- TradingView API root: `window.TradingViewApi` (объект инжектируется TradingView Desktop).
- Exit codes: 0 (success), 1 (error), 2 (connection failure).

### Pending

- P0.02 Зафиксировать commit hash оригинального репозитория.
- P0.03–P0.04 Запустить `npm install` и `npm test`.
- P2 MCP stdio server (JSON-RPC types уже созданы; нужны доп. тесты).
- P3 CDP client (WebSocket + Runtime.evaluate реализованы; нужны тесты с mock-сервером).
- P4 E2E: `tv_health_check`, `tv_discover`, `tv_ui_state`, `tv_launch`, CLI `tv status`.
- P2W.04–P2W.12 CLI flag --tv-path, doctor windows, unit-тесты discovery.

---

## 2026-04-24 (P1 — Go skeleton)

### Added — Go skeleton files

- `go.mod` — модуль `github.com/jhonroun/tradingview-mcp-go`, Go 1.21, зависимость `github.com/gorilla/websocket v1.5.3`.
- `internal/mcp/types.go` — JSON-RPC 2.0 и MCP protocol типы: Request, Response, RPCError, Tool, InputSchema, ListToolsResult, CallToolResult.
- `internal/mcp/registry.go` — ToolDef и Registry: Register, Get, List, Call.
- `internal/mcp/server.go` — MCP stdio server: bufio read loop, handleInitialize, handleListTools, handleCallTool, handlePing.
- `internal/mcp/server_test.go` — 5 unit-тестов: initialize, list tools, call tool, unknown method, registry call unknown.
- `internal/cdp/types.go` — Target, Message, CDPError, EvaluateParams, RemoteObject, EvaluateResult, ExceptionDetails.
- `internal/cdp/discovery.go` — ListTargets (GET /json/list), FindChartTarget (prefer chart URL, fallback to tradingview URL).
- `internal/cdp/discovery_test.go` — 5 unit-тестов: exact match, fallback, none, skip non-page, prefer chart over root.
- `internal/cdp/client.go` — Client: Connect, ConnectWithRetry (exponential backoff), EnableDomains, Evaluate, LivenessCheck, Close; goroutine-safe request/response matching.
- `internal/tools/health/health.go` — HealthCheck, Discover, UIState, Launch; RegisterTools → регистрирует tv_health_check, tv_discover, tv_ui_state, tv_launch.
- `internal/discovery/discovery.go` — Find: TRADINGVIEW_PATH env → LOCALAPPDATA/PROGRAMFILES → Microsoft Store via PowerShell Get-AppxPackage → /Applications (macOS) → PATH (Linux).
- `internal/launcher/launcher.go` — Launch: killRunning, exec TradingView с --remote-debugging-port, ожидание CDP с таймаутом 15 с.
- `internal/cli/router.go` — Register, Dispatch, parseFlags (--key=value, --key value, --bool-flag).
- `cmd/tvmcp/main.go` — MCP stdio сервер: регистрирует health tools, запускает Server.Run().
- `cmd/tv/main.go` — CLI: команды `status` и `launch`.

### Verified

- `go build ./...` — успешно, нет ошибок.
- `go test ./...` — 10/10 тестов PASS (5 cdp/discovery, 5 mcp/server).

---

## 2026-04-24 (P2/P3/P4 — MCP server, CDP client, first E2E tools)

### Added — P3 CDP enhancements

- `internal/cdp/client.go` — добавлен `CaptureScreenshot` (Page.captureScreenshot → base64 PNG).
- `internal/cdp/client_test.go` — 5 тестов через mock WebSocket CDP сервер: Evaluate (число), EnableDomains, LivenessCheck, JS-ошибка, Screenshot.

### Added — CLI commands

- `cmd/tv/main.go` — команды `discover` (tv_discover), `ui-state` (tv_ui_state), `doctor` (диагностика CDP + installation).
- `tv doctor` выводит JSON с ключами `cdp` (ok, targets, chart, targetId, url) и `install` (ok, path, source, platform).

### Added — unit tests

- `internal/cli/router_test.go` — 5 тестов parseFlags: equals-form, space-form, bool-flag, empty, mixed.
- `internal/discovery/discovery_test.go` — 3 теста: TRADINGVIEW_PATH env, missing env path, fileExists helper.

### Verified — P2/P3/P4

---

## 2026-04-24 (P5 — Read-only chart tools)

### Added — P5 packages

- `internal/cdp/session.go` — `WithSession` helper: connect → enable domains → run fn → close. Eliminates boilerplate in every tool.
- `internal/cdp/client.go` — `ScreenshotClip` type, `CaptureScreenshotClip(clip)`, refactored `screenshot()` private method.
- `internal/tradingview/js.go` — `ChartAPI`, `BarsPath`, `ChartWidget` constants; `SafeString()` (mirrors connection.js safeString).
- `internal/tools/chart/chart.go` — `chart_get_state` (symbol, resolution, chartType, studies list), `chart_get_visible_range`.
- `internal/tools/data/data.go` — 12 tools: `data_get_ohlcv` (with Go-side summary computation), `quote_get`, `data_get_study_values`, `data_get_pine_lines` (horizontal-level dedup + verbose), `data_get_pine_labels` (max_labels + verbose), `data_get_pine_tables` (row formatting), `data_get_pine_boxes` (zone dedup + verbose), `data_get_indicator`, `data_get_strategy_results`, `data_get_trades`, `data_get_equity`, `depth_get`.
- `internal/tools/capture/capture.go` — `capture_screenshot` (regions: full / chart / strategy_tester; JS clip detection; saves to screenshots/).
- `cmd/tvmcp/main.go` — registers chart, data, capture tool groups (total ~18 new MCP tools).
- `cmd/tv/main.go` — CLI commands: `tv quote [SYMBOL]`, `tv ohlcv [--count N] [--summary]`, `tv screenshot [--region X] [--filename F]`, `tv chart-state`.

### Added — P5 tests

- `internal/tools/data/data_test.go` — TestRegisteredToolNames (12 tool names vs compatibility matrix), TestRound2, TestBuildGraphicsJSContainsFilter, TestJoinStr, TestCoalesce.
- `internal/tradingview/js_test.go` — TestSafeString (5 cases), TestSafeStringNoInjection (3 dangerous inputs).

### Pending — P5

- P5.16 `batch_run` — complex multi-symbol iteration, deferred to later session.

### Verified — P5 build and tests

- `go build ./...` — успешно.
- `go test ./...` — **30/30 тестов PASS** (5 cdp/client, 5 cdp/discovery, 5 cli/router, 3 discovery, 5 mcp/server, 5 tools/data, 2 tradingview/js).

- `go build ./...` — успешно.
- `go test ./...` — **23/23 тестов PASS** (5 cdp/client, 5 cdp/discovery, 5 cli/router, 3 discovery, 5 mcp/server).

---

## 2026-04-24 (P6 — Chart control + indicator tools)

### Added — P6 packages

- `internal/tools/chart/wait.go` — `waitForChartReady`: polls every 200 ms; checks loading spinner, symbol match (case-insensitive), bar count stable ×2; timeout 10 s.
- `internal/tools/chart/control.go` — `SetSymbol` (Promise + waitForChartReady), `SetTimeframe` (setResolution), `SetType` (chartTypeMap 0–9), `ManageIndicator` (add: createStudy + diff, remove: removeEntity), `SetVisibleRange` (bar-index scan + zoomToBarsRange), `ScrollToDate` (±25 bar window); `registerControlTools` adds 6 MCP tools.
- `internal/tools/chart/symbol.go` — `SymbolInfo` (chart.symbolExt()), `SymbolSearch` (REST GET symbol-search.tradingview.com, strips `<em>`, max 15 results); `registerSymbolTools` adds 2 MCP tools.
- `internal/tools/indicators/indicators.go` — `SetInputs` (getInputValues → override by id → setInputValues), `ToggleVisibility` (setVisible + isVisible); `RegisterTools` adds 2 MCP tools.
- `internal/tools/chart/chart.go` — `RegisterTools` now calls `registerControlTools` + `registerSymbolTools` (total 10 chart tools).
- `cmd/tvmcp/main.go` — registers `indicators.RegisterTools` (12 total indicator + 10 chart tools).
- `cmd/tv/main.go` — CLI commands: `tv set-symbol`, `tv set-timeframe`, `tv set-type`, `tv symbol-info`, `tv symbol-search [--type] [--exchange]`, `tv indicator-toggle ENTITY_ID [--visible]`.

### Added — P6 tests

- `internal/tools/chart/chart_p6_test.go` — TestChartTypeMap (9 entries), TestSetTypeUnknown, TestSetTypeNormalization, TestRegisterToolsP6Names (10 tool names).
- `internal/tools/indicators/indicators_test.go` — TestRegisterIndicatorTools (2 tool names).

### Verified — P6 build and tests

- `go build ./...` — успешно.
- `go test ./...` — **38/38 тестов PASS** (5 cdp/client, 5 cdp/discovery, 5 cli/router, 3 discovery, 5 mcp/server, 5 tools/data, 4 tools/chart, 1 tools/indicators, 2 tradingview/js).

---

## 2026-04-24 (P7 — Pine Script tools)

### Added — P7 packages

- `internal/cdp/client.go` — `KeyEventParams`, `DispatchKeyEvent(ctx, params)` for Ctrl+Enter / Ctrl+S dispatch.
- `internal/tools/pine/pine.go` — all CDP-dependent Pine tools:
  - `findMonaco` constant — React Fiber traversal to locate Monaco editor instance.
  - `ensurePineEditorOpen(ctx, client)` — opens Pine Editor panel, polls until Monaco ready (10 s).
  - `GetSource()` — `editor.getValue()`.
  - `SetSource(source)` — `editor.setValue(escaped)`.
  - `Compile()` — clicks "Save and add to chart" / "Add to chart" / "Update on chart" buttons; fallback Ctrl+Enter; waits 2 s.
  - `SmartCompile()` — counts studies before/after, clicks compile, reads Monaco markers, reports study_added.
  - `GetErrors()` — `getModelMarkers({resource: model.uri})`.
  - `GetConsole()` — DOM scraping for console rows with timestamp/type classification.
  - `Save()` — Ctrl+S + handles "Save Script" name dialog.
  - `NewScript(type)` — injects indicator/strategy/library template via `editor.setValue`.
  - `OpenScript(name)` — fetch pine-facade list (credentials:include) + fuzzy name match + get source + Monaco setValue.
  - `ListScripts()` — fetch pine-facade list, returns id/name/title/version/modified.
  - `Check(source)` — HTTP POST to `pine-facade.tradingview.com/pine-facade/translate_light` (Guest, public endpoint); returns errors/warnings.
- `internal/tools/pine/analyze.go` — offline static analyzer (no CDP):
  - Detects Pine version from `//@version=N`.
  - Tracks `array.from()` and `array.new*()` declarations with sizes.
  - Flags `array.get/set` calls with literal out-of-bounds indices.
  - Flags `array.first/last()` on zero-size arrays.
  - Flags `strategy.entry/close` without `strategy()` declaration.
  - Warns about Pine versions < 5.
- `cmd/tvmcp/main.go` — registers `pine.RegisterTools(reg)` (12 new MCP tools; total ~42).
- `cmd/tv/main.go` — `tv pine <get|set|compile|smart-compile|raw-compile|errors|console|save|new|open|list|analyze|check>` dispatcher.

### Added — P7 tests

- `internal/tools/pine/pine_test.go` — TestRegisterPineToolNames (12 names), TestAnalyzeCleanScript, TestAnalyzeArrayOutOfBounds, TestAnalyzeStrategyWithoutDecl, TestAnalyzeOldVersion, TestAnalyzeStrategyWithDecl.

### Verified — P7 build and tests

- `go build ./...` — успешно.
- `go test ./...` — **43/43 тестов PASS** (5 cdp/client, 5 cdp/discovery, 5 cli/router, 3 discovery, 5 mcp/server, 5 tools/data, 4 tools/chart, 1 tools/indicators, 5 tools/pine, 2 tradingview/js).

---

## 2026-04-24 (P8 — Drawing tools)

### Added — P8 packages

- `internal/tools/drawing/drawing.go` — 5 MCP tools:
  - `DrawShape` — `createShape` (1-point) or `createMultipointShape` (2-point); waits 200 ms; diffs `getAllShapes()` to extract new entity ID. `fmtNum` uses `strconv.FormatFloat('f')` to avoid scientific notation in JS. `requireFinite` validates point coordinates.
  - `ListDrawings` — `getAllShapes()` → `[{id, name}]`.
  - `GetProperties` — `getShapeById(eid)` → points, properties (with `.properties()` fallback), visibility, lock, selection state, available methods.
  - `RemoveOne` — pre-checks shape exists, calls `removeEntity(eid)`, post-verifies removal; returns `remaining_shapes`.
  - `ClearAll` — `removeAllShapes()`.
- `cmd/tvmcp/main.go` — registers `drawing.RegisterTools(reg)` (5 new MCP tools; total ~47).
- `cmd/tv/main.go` — `tv draw <shape|list|get|remove|clear>` dispatcher; `tv draw shape` accepts `--time/--price/--time2/--price2/--text` flags.

### Added — P8 tests

- `internal/tools/drawing/drawing_test.go` — TestRegisterDrawingToolNames (5 names), TestRequireFinite (4 cases), TestFmtNum (4 cases), TestDrawShapeValidation (NaN guard).

### Verified — P8 build and tests

- `go build ./...` — успешно.
- `go test ./...` — **47/47 тестов PASS** (5 cdp, 5 cdp/discovery, 5 cli/router, 3 discovery, 5 mcp/server, 5 tools/data, 4 tools/chart, 4 tools/drawing, 1 tools/indicators, 5 tools/pine, 2 tradingview/js).

---

## 2026-04-24 (P9 — Alerts + Watchlist)

### Added — P9 packages

- `internal/cdp/client.go` — `InsertText(ctx, text)` via `Input.insertText`.
- `internal/tools/alerts/alerts.go` — 5 MCP tools (alert + watchlist group):
  - `CreateAlert` — clicks "Create Alert" button or Shift+A fallback; sets price via React synthetic event override (`Object.getOwnPropertyDescriptor(HTMLInputElement.prototype,'value').set`); sets message via textarea; clicks "Create" button.
  - `ListAlerts` — fetch `pricealerts.tradingview.com/list_alerts` (credentials:include) via CDP; returns alert_id, symbol, type, condition, active, timestamps.
  - `DeleteAlerts` — `delete_all:true`: opens context menu (manual confirmation required); `delete_all:false`: returns "not yet supported" error (matches Node.js behavior).
  - `GetWatchlist` — three-tier DOM scraping: `data-symbol-full` attributes → `symbolName/tickerName` text scan; returns symbol + last/change/change_percent.
  - `AddToWatchlist` — opens watchlist panel, clicks add-symbol button, `InsertText(symbol)`, Enter to confirm, Escape to close; Escape cleanup on error.
- `cmd/tvmcp/main.go` — registers `alerts.RegisterTools(reg)` (5 new tools; total ~52).
- `cmd/tv/main.go` — `tv alert <list|create|delete>`, `tv watchlist <get|add SYMBOL>`.

### Added — P9 tests

- `internal/tools/alerts/alerts_test.go` — TestRegisterAlertToolNames (5 names), TestDeleteAlertsIndividualError.

### Verified — P9 build and tests

- `go build ./...` — успешно.
- `go test ./...` — **49/49 тестов PASS** (5 cdp, 5 cdp/discovery, 5 cli/router, 3 discovery, 5 mcp/server, 2 tools/alerts, 5 tools/data, 4 tools/chart, 4 tools/drawing, 1 tools/indicators, 5 tools/pine, 2 tradingview/js).

---

## 2026-04-24 (P10 — Panes + Tabs)

### Added — P10 packages

- `internal/tools/pane/pane.go` — 4 MCP tools:
  - `ListPanes` — evaluates `_chartWidgetCollection.getAll()`, returns layout code/name, chart_count, active_index, per-pane symbol/resolution.
  - `SetLayout` — `resolveLayout` normalises aliases (single→s, 2x2→4, quad→4, grid→4, 2x1→2h, 1x2→2v) and friendly names; calls `_chartWidgetCollection.setLayout(code)`; waits 500 ms; returns updated pane list.
  - `FocusPane` — finds pane by 0-based index in `getAll()`, calls `_mainDiv.click()`; returns focused_index and total_panes.
  - `SetPaneSymbol` — focuses pane first (FocusPane + 300 ms), then calls `chart.setSymbol(symbol, {})` via Promise (500 ms internal delay).
- `internal/tools/tab/tab.go` — 4 MCP tools:
  - `ListTabs` — GET `/json/list`; filters TradingView pages; returns id, title, url per tab.
  - `NewTab` — sends Ctrl+T (modifiers=2, keyCode=84) to active window; waits 1 s; returns updated tab list.
  - `CloseTab` — guards against closing the last tab (≥2 required); sends Ctrl+W; waits 500 ms; returns updated tab list.
  - `SwitchTab(tabID)` — GET `http://localhost:9222/json/activate/{id}`; returns activated_tab_id.
- `cmd/tvmcp/main.go` — registers `pane.RegisterTools` + `tab.RegisterTools` (8 new tools; total ~60).
- `cmd/tv/main.go` — `tv pane <list|set-layout|focus|set-symbol>` and `tv tab <list|new|close|switch ID>` dispatchers.

### Fixed — P10

- `internal/tools/pane/pane.go` — typo `mpc.PropertySchema` → `mcp.PropertySchema` in `pane_focus` registration.

### Added — P10 tests

- `internal/tools/pane/pane_test.go` — TestResolveLayoutKnownCodes (4 codes), TestResolveLayoutAliases (8 aliases), TestResolveLayoutCaseInsensitive, TestResolveLayoutUnknown, TestRegisterPaneToolNames (4 names).
- `internal/tools/tab/tab_test.go` — TestRegisterTabToolNames (4 names), TestSwitchTabEmptyID.

### Verified — P10 build and tests

- `go build ./...` — успешно.
- `go test ./...` — **56/56 тестов PASS** (5 cdp, 5 cdp/discovery, 5 cli/router, 3 discovery, 5 mcp/server, 2 tools/alerts, 5 tools/data, 4 tools/chart, 4 tools/drawing, 1 tools/indicators, 5 tools/pane, 5 tools/pine, 2 tools/tab, 2 tradingview/js).

---

## 2026-04-24 (P11 — Replay)

### Added — P11 packages

- `internal/tools/replay/replay.go` — 6 MCP tools (1:1 port of `src/core/replay.js`):
  - `Start(date)` — checks `isReplayAvailable()`, shows toolbar, calls `selectDate(tsMs)` (awaited) or `selectFirstAvailableDate()`, polls 30×250 ms until `isReplayStarted && currentDate != null`; on failure calls `stopReplay()` and returns descriptive error.
  - `Step()` — guards `isReplayStarted`, reads `currentDate` before, calls `doStep()`, polls 12×250 ms until date changes.
  - `Stop()` — idempotent: returns `already_stopped` if not running, else calls `stopReplay()`.
  - `Status()` — single JS block reads `isReplayAvailable/Started/AutoplayStarted, replayMode, currentDate, autoplayDelay`; appends `position` and `realizedPL`.
  - `Autoplay(speedMs)` — validates speed against `VALID_AUTOPLAY_DELAYS` (100,143,200,300,1000,2000,3000,5000,10000) **before** CDP calls; calls `changeAutoplayDelay` if non-zero, then `toggleAutoplay`; returns `autoplay_active` + `delay_ms`.
  - `Trade(action)` — dispatches `buy()`, `sell()`, `closePosition()`; returns `position` + `realized_pnl`.
  - `wv(path)` helper — unwraps TradingView observable (mirrors Node.js `wv()` in core/replay.js).
- `cmd/tvmcp/main.go` — registers `replay.RegisterTools(reg)` (6 new tools; total ~66).
- `cmd/tv/main.go` — `tv replay <start [--date YYYY-MM-DD]|step|stop|status|autoplay [--speed MS]|trade buy|sell|close>` dispatcher.

### Added — P11 tests

- `internal/tools/replay/replay_test.go` — TestRegisterReplayToolNames (6 names + count), TestAutoplayInvalidSpeed, TestAutoplayValidSpeeds (9 valid delays), TestTradeInvalidAction, TestWvHelper.

### Verified — P11 build and tests

- `go build ./...` — успешно.
- `go test ./...` — **60/60 тестов PASS** (5 cdp, 5 cdp/discovery, 5 cli/router, 3 discovery, 5 mcp/server, 2 tools/alerts, 5 tools/data, 4 tools/chart, 4 tools/drawing, 1 tools/indicators, 5 tools/pane, 5 tools/pine, 4 tools/replay, 2 tools/tab, 2 tradingview/js).

---

## 2026-04-25 (P12 — UI automation + Layouts)

### Added — P12 packages

- `internal/cdp/client.go` — `MouseEventParams` struct and `DispatchMouseEvent(ctx, p)` via `Input.dispatchMouseEvent`.
- `internal/tools/ui/ui.go` — 12 MCP tools (1:1 port of `src/core/ui.js`):
  - `Click(by, value)` — finds element by aria-label / data-name / text / class-contains; calls `.click()`; returns tag, text, aria_label, data_name.
  - `OpenPanel(panel, action)` — bottom panels (pine-editor, strategy-tester) use `bottomWidgetBar.activateScriptEditorTab/showWidget/hideWidget`; side panels (watchlist, alerts, trading) use data-name button with aria-pressed state detection; action: open/close/toggle.
  - `Fullscreen()` — clicks `[data-name="header-toolbar-fullscreen"]`.
  - `LayoutList()` — `getSavedCharts` Promise (awaitPromise=true, 5 s timeout); returns id/name/symbol/resolution/modified per layout.
  - `LayoutSwitch(name)` — numeric ID → `loadChartFromServer(id)` directly; name → `getSavedCharts` exact match then substring match → `loadChartFromServer`; waits 500 ms; dismisses "unsaved changes" dialog (open anyway / don't save / discard) if present; waits 1 s after dismiss.
  - `Keyboard(key, modifiers)` — keyMap for 17 named keys; fallback to `Key<UPPER>` + charCodeAt(0); modifiers: alt=1, ctrl=2, meta=4, shift=8; dispatches keyDown + keyUp.
  - `TypeText(text)` — `Input.insertText`; returns typed (capped 100 chars) + length.
  - `Hover(by, value)` — finds element coords via `getBoundingClientRect()` centre; dispatches `mouseMoved`.
  - `Scroll(direction, amount)` — finds chart canvas centre; dispatches `mouseWheel` with deltaX/deltaY; default 300 px.
  - `MouseClick(x, y, button, doubleClick)` — `mouseMoved` → `mousePressed` → `mouseReleased`; optional 50 ms + second press/release for double click.
  - `FindElement(query, strategy)` — text: textContent scan on interactive elements (max 20, visible only); aria-label: `[aria-label*=query]`; css: `querySelectorAll(query)`; returns tag/text/aria_label/data_name/x/y/width/height/visible per element.
  - `Evaluate(expression)` — raw `Runtime.evaluate` passthrough; returns `{ success, result }`.
- `cmd/tvmcp/main.go` — registers `ui.RegisterTools(reg)` (12 new tools; total ~78).
- `cmd/tv/main.go` — `tv ui <click|open-panel|fullscreen|keyboard|type|hover|scroll|mouse|find|eval>` and `tv layout <list|switch NAME>` dispatchers.

### Added — P12 tests

- `internal/tools/ui/ui_test.go` — TestRegisterUIToolNames (12 names + count), TestKeyMapEntries (17 keys), TestKeyboardModifierBitfield (7 cases), TestScrollDefaultAmount, TestMouseClickButtonNormalise (5 cases), TestFindElementDefaultStrategy.

### Verified — P12 build and tests

- `go build ./...` — успешно.
- `go test ./...` — **66/66 тестов PASS** (5 cdp, 5 cdp/discovery, 5 cli/router, 3 discovery, 5 mcp/server, 2 tools/alerts, 5 tools/data, 4 tools/chart, 4 tools/drawing, 1 tools/indicators, 5 tools/pane, 5 tools/pine, 4 tools/replay, 2 tools/tab, 6 tools/ui, 2 tradingview/js).

---

## 2026-04-25 (P13 — Streaming)

### Added — P13 packages

- `internal/stream/stream.go` — JSONL streaming engine (1:1 port of `src/core/stream.js`):
  - `pollLoop(ctx, w, errW, label, intervalMs, dedupe, fetcher)` — connects once; reuses single CDP client; on CDP/WebSocket error reconnects after 2 s; dedup via `JSON.stringify` comparison; appends `_ts` (UnixMilli) and `_stream` fields to every emitted line; exits cleanly on `ctx.Done()`.
  - `isCDPError(err)` — detects CDP/websocket/connection errors for silent reconnect.
  - `StreamQuote(ctx, w, errW, intervalMs)` — last bar OHLCV; default 300 ms.
  - `StreamBars(ctx, w, errW, intervalMs)` — last bar with symbol/resolution/bar_index; default 500 ms.
  - `StreamValues(ctx, w, errW, intervalMs)` — all visible indicator `_lastBarValues`; default 500 ms.
  - `StreamLines(ctx, w, errW, intervalMs, filter)` — Pine `line.new()` price levels, deduped + sorted desc; default 1000 ms.
  - `StreamLabels(ctx, w, errW, intervalMs, filter)` — Pine `label.new()` text+price, max 50; default 1000 ms.
  - `StreamTables(ctx, w, errW, intervalMs, filter)` — Pine `table.new()` row data; default 2000 ms.
  - `StreamAllPanes(ctx, w, errW, intervalMs)` — all chart panes OHLCV in one tick; default 500 ms.
- `cmd/tv/main.go` — `tv stream` handled before `cli.Dispatch` (streams never return):
  - `signal.NotifyContext` for SIGINT/SIGTERM graceful shutdown.
  - Subcommands: `quote bars values lines labels tables all`.
  - `--interval MS` and `--filter NAME` flags parsed inline.
  - Compliance notice printed to stderr before any stream starts.

### Added — P13 tests

- `internal/stream/stream_test.go` — TestDefaultIntervals (7 streams × default + explicit), TestBuildLinesExprContainsFilter, TestBuildLabelsExprContainsFilter, TestBuildTablesExprContainsFilter, TestIsCDPError (7 cases), TestPollLoopCancelledContext (exit on pre-cancelled ctx + "stopped" in stderr), TestJSONLTimestampFields (`_ts` + `_stream` present), TestConstantsContainTradingViewAPI.

### Verified — P13 build and tests

- `go build ./...` — успешно.
- `go test ./...` — **73/73 тестов PASS** (5 cdp, 5 cdp/discovery, 5 cli/router, 3 discovery, 5 mcp/server, 2 tools/alerts, 5 tools/data, 4 tools/chart, 4 tools/drawing, 1 tools/indicators, 5 tools/pane, 5 tools/pine, 4 tools/replay, 7 stream, 2 tools/tab, 6 tools/ui, 2 tradingview/js).

---

## 2026-04-25 (P14 — Compatibility audit + batch_run + README)

### Added — batch_run (P5.16 backlog)

- `internal/tools/batch/batch.go` — `batch_run` MCP tool (1:1 port of `src/core/batch.js`):
  - Iterates every `symbol × timeframe` combination.
  - Per iteration: `chart.SetSymbol` (includes waitForChartReady), optional `chart.SetTimeframe` + 500 ms settle, user `delay_ms` (default 2000 ms).
  - Actions: `screenshot` (capture.CaptureScreenshot "chart" region, safe filename), `get_ohlcv` (exportData Promise, cap 500 bars), `get_strategy_results` (DOM scrape backtesting panel + 1 s settle).
  - Returns `{ success, total_iterations, successful, failed, results[] }`.
- `cmd/tvmcp/main.go` — registers `batch.RegisterTools(reg)` (total **78 MCP tools**).
- `cmd/tv/main.go` — `tv batch --symbols SYM1,SYM2 --action ACTION [--timeframes TF1,TF2] [--delay MS] [--count N]`.

### Compatibility audit results (P14.01–P14.03)

- **MCP tools**: Go 78 / Node.js 78 — exact match, zero missing, zero extra.
- **CLI groups**: all 15 Node.js groups covered (health, chart, data, capture, indicators, pine, drawing, pane, replay, tab, alerts, ui, layout, watchlist, stream); `batch` added as bonus CLI command.
- **Tool names**: verified via live `tools/list` RPC against sorted Node.js grep — 100% match.
- **JSON output structure**: `{ success: bool, ...fields }` pattern preserved across all tools.

### Added — README.md (P14.08)

- `README.md` — user-facing documentation: requirements, build, launch, Claude Code MCP config, full tool table (78 tools grouped), CLI reference, architecture diagram, compatibility notes, disclaimer.

### Added — P14 tests

- `internal/tools/batch/batch_test.go` — TestRegisterBatchToolName, TestBatchRunEmptySymbols, TestBatchRunDefaultDelayAndCount, TestBatchRunOhlcvCountCap, TestBatchRunTimeframeDefault.

### Verified — P14 build and tests

- `go build ./...` — успешно.
- `go test ./...` — **78/78 тестов PASS** (5 tools/batch + all prior).

---

## 2026-04-25 (P15 — Build system, Skills, Documentation)

### Added — P2W.04: --tv-path CLI flag

- `internal/launcher/launcher.go` — `Launch(port, killExisting, tvPath string)`: when `tvPath` is non-empty, skips auto-discovery and uses the provided path directly; validates file existence; reports `source: "cli-flag"` in response.
- `internal/tools/health/health.go` — `LaunchArgs.TvPath *string` field; MCP `tv_launch` schema updated with `tv_path` property; `Launch()` threads `tvPath` to launcher.
- `cmd/tv/main.go` — `tv launch --tv-path=PATH` parses `opts["tv-path"]` and populates `LaunchArgs.TvPath`.

### Added — Build system (P15.01–P15.04)

- `Makefile` — targets: `build` (current platform → `bin/`), `build-all` (6 platforms: windows/linux/darwin × amd64/arm64), `install` (`go install`), `test`, `test-verbose`, `clean`, `release` (build-all + ZIP/tar.gz archives in `bin/releases/`).
- `scripts/build.sh` — shell build script; respects `GOOS`/`GOARCH` env vars.
- `scripts/build.bat` — Windows batch build script.
- `scripts/install.sh` — installs to `/usr/local/bin` (or `PREFIX`).
- `scripts/install.bat` — installs to `%SystemRoot%\System32` (or first argument).

### Added — Scripts (P0S.05–P0S.07)

- `scripts/pine_pull.sh` + `scripts/pine_pull.bat` — wraps `tv pine get > scripts/current.pine`; drop-in replacement for original `pine_pull.js`.
- `scripts/pine_push.sh` + `scripts/pine_push.bat` — wraps `tv pine set` + `tv pine smart-compile`; drop-in replacement for original `pine_push.js`.
- `scripts/launch_tv_debug.bat`, `scripts/launch_tv_debug.vbs`, `scripts/launch_tv_debug_linux.sh`, `scripts/launch_tv_debug_mac.sh` — copied from original Node.js project; still functional as standalone launchers alongside `tv launch`.

### Added — Skills (P0S.08)

- `skills/chart-analysis/SKILL.md` — technical analysis workflow (unchanged from original; no Node.js script references).
- `skills/multi-symbol-scan/SKILL.md` — multi-symbol scan using `batch_run`; JSON input examples instead of code blocks.
- `skills/pine-develop/SKILL.md` — Pine development loop updated to reference `tv pine` CLI and `scripts/pine_pull.sh` / `pine_push.sh` instead of `node scripts/pine_pull.js`.
- `skills/replay-practice/SKILL.md` — replay practice workflow (unchanged from original).
- `skills/strategy-report/SKILL.md` — strategy report workflow (unchanged from original).

### Added — Documentation (P15.05–P15.08)

- `docs/ru/README.md` — primary full Russian documentation: история портирования (Node.js → Go с Claude Code), требования, сборка, запуск, MCP-конфиг, таблица 78 инструментов, CLI-справка, скрипты, навыки, архитектура, совместимость, дисклеймер.
- `docs/en/README.md` — full English documentation: same structure including Origin Story section (porting process, Claude Code role).
- `README.md` (корень) — обновлён: навигационная таблица → docs/ru + docs/en, раздел поддерживаемых MCP-клиентов (не привязан к Claude Code), история портирования, layout структуры проекта.

### Changed — README.md root

- Rewritten as a concise navigation hub pointing to `docs/ru/README.md` and `docs/en/README.md` for full content.
- Added "Supported MCP Clients" section: Claude Code, Cursor, Cline, Continue, Windsurf, any stdio MCP client.
- Added project history note: Go port performed with Claude Code assistance, April 2026.

### Verified — P15 build and tests

- `go build ./...` — успешно.
- `go test ./...` — **78/78 тестов PASS** (все предыдущие; новые файлы не имеют тестируемого Go-кода).
- `bash scripts/build.sh` — `bin/tvmcp.exe` + `bin/tv.exe` — успешно.
