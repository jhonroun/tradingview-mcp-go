# Final Audit Report - HTS Readiness

Date: 2026-04-27

Status: `GO WITH LIMITATIONS`

Scope: final audit after repository inspection, unit tests, live smoke tests,
CDP checks, TradingView data-source research, Pine/strategy/replay live audit,
MOEX bid/ask research, and HTS contract design.

Boundary: this repository remains a TradingView Desktop MCP/CLI implementation.
HTS decision logic, Tinkoff integration, broker execution, account/risk logic,
and optimizer orchestration belong to an external HTS MCP layer.

---

## Executive Summary

`tradingview-mcp-go` is usable as a working TradingView MCP base for HTS
integration, but it is not a complete numerical trading-data layer by itself.

Confirmed ready:

- MCP registry and stdio protocol are operational.
- CLI commands are implemented and useful for debugging.
- CDP connection to TradingView Desktop works.
- Core chart context, OHLCV, screenshots, symbol switching, symbol search, and
  continuous contract context are usable.
- MCP stdin reader no longer has the historical `bufio.Scanner` 64 KiB risk.
- Screenshot filename `.png.png` bug is fixed and unit-tested.
- Source-aware HTS/LLM contracts are documented.

Confirmed limitations:

- Indicator values returned by current public indicator tools are not safe for
  HTS numerical decisions until locale parsing and source flags are fixed.
- Real indicator history exists in TradingView internal study model
  `study.data().valueAt(index)`, but current public tools do not expose it.
- Strategy results/trades/equity tools still return partial or empty success
  states when no strategy is loaded.
- MOEX futures bid/ask are unavailable in the tested TradingView quote state.
  Executable bid/ask/spread should come from Tinkoff orderbook.
- Pine compile/open/save workflows can mutate the current Pine editor or saved
  scripts and must be treated as high-risk.

Decision:

```text
GO WITH LIMITATIONS
```

HTS can build on this MCP for TradingView chart context, OHLCV, visual evidence,
Pine source access when permitted, and operator workflows. HTS must calculate
critical features in Go or consume verified Pine `HTS_JSON`, and must use
Tinkoff for executable instruments, bid/ask/spread, expiration, margin, and
trading status.

---

## Readiness Matrix

| Area | Status | Evidence | HTS Use |
| --- | --- | --- | --- |
| MCP registry | ready | 82 tools registered: 78 original parity + 4 HTS composite | Safe |
| MCP stdio | ready | `Reader.ReadBytes('\n')` + 16 MB guard; large-message tests | Safe |
| CLI | ready | 29 registered commands + stream dispatcher | Safe for diagnostics |
| CDP connection | ready/live-tested | TradingView Desktop reachable via CDP | Safe |
| Chart state | ready/live-tested | `chart_get_state` smoke OK | Safe |
| Symbol switching | ready/live-tested | `chart_set_symbol` tested in prior smoke workflows | Safe with operator awareness |
| OHLCV | ready/live-tested | `data_get_ohlcv` smoke OK on `RUS:NG1!` | Safe as TradingView bars |
| Quote last/OHLC | partial | `quote_get` works for last/OHLCV | Safe for last/OHLC, unsafe for absent BBO |
| MOEX bid/ask | unavailable in TV | `RUS:NG1!` and `RUS:NGJ2026` lack bid/ask keys | Use Tinkoff |
| Symbol search | fixed/regression watch | API params fixed; tests and CLI checks passed | Safe with endpoint watch |
| Screenshot | ready/unit-tested | `.png.png` fixed; non-PNG ext rejected | Safe for visual evidence |
| Indicator current values | unsafe | Data Window text parser breaks decimal comma and scale suffixes | Not safe for numerical decisions |
| Indicator history | source found/not exposed | Internal `study.data().valueAt(index)` works in research | Implement next |
| Pine source read | partial/live-tested | `pine_get_source` returned current editor source | Safe when source is visible to user |
| Pine source mutation | high risk | `pine_compile` saved over active script during audit | Operator-only |
| Pine `HTS_JSON` | recommended pattern | Documented as machine-readable output path | Safe when schema/hash verified |
| Strategy metrics | partial/unsafe | empty metrics/trades/equity can appear as success | Not safe until status contract fixed |
| Replay | partial/live-tested | replay actions work; trade state weak | Safe for practice, not execution truth |
| HTS composite tools | partial | useful, but depend on unsafe indicator values | Use only with quality flags |
| LLM context contracts | documented/draft | market summary, LLM context, resolver contracts added | Ready for external HTS design |
| Tinkoff resolver | documented/draft | external resolver contract added | Implement outside this repo |

---

## Ready For HTS

These capabilities can be used now with normal provenance and freshness checks:

- TradingView Desktop CDP connectivity and diagnostics.
- MCP protocol basics: initialize, tools/list, tools/call, ping, error handling.
- Chart state: symbol, ticker, exchange, timeframe, chart type, pane count,
  loaded studies metadata.
- OHLCV bars from TradingView, with the caveat that they are TradingView/feed
  bars and not broker execution data.
- Current last/OHLC/volume quote fields when present.
- Symbol search after the query-param fix, with regression watch.
- Screenshots for visual audit, operator review, reports, and LLM visual
  context when explicitly requested.
- Pine source read when TradingView already exposes source to the current user.
- Pine script list/read paths for user-owned scripts, with mutation caveats for
  open/write workflows.
- Continuous futures context as analysis metadata only.
- Indicator input updates for applied indicators, with before/after verification.
- MCP/CLI as a data-collection layer for an external HTS summary builder.
- Draft contracts:
  - `HTS_MARKET_SUMMARY_CONTRACT.md`;
  - `LLM_MARKET_CONTEXT_CONTRACT.md`;
  - `INSTRUMENT_RESOLVER_CONTRACT.md`.

---

## Ready Only For Visual Analysis

These are useful for human review, screenshots, UI automation, and explanatory
context, but not standalone numerical truth:

- `capture_screenshot`.
- DOM/legend/Data Window text snapshots.
- Pine labels, lines, boxes, and tables unless they follow strict `HTS_JSON`.
- Replay UI state and replay screenshots.
- Chart drawings and UI interactions.
- Pine console/log scrape.
- `market_summary` and `chart_context_for_llm` when they include current
  indicator values from unsafe sources.

---

## Not Safe For Numerical Decisions

Do not feed these to HTS or LLM as factual numerical inputs:

- Canvas coordinates, screenshot pixels, or any Y-coordinate-derived value.
- Current `data_get_indicator` / `data_get_study_values` numeric values before
  locale parsing and source flags are fixed.
- Data Window values parsed from localized strings such as `31,51`, `14,63 K`,
  or similar display values.
- TradingView `bid=0` / `ask=0` sentinels for MOEX futures.
- TradingView MOEX bid/ask when internal quote keys are absent.
- Empty strategy metrics, empty trades, or empty equity returned with
  `success:true`.
- Pine compile success unless chart studies, compiler markers, and mutation
  side effects are verified.
- Continuous futures symbols as execution symbols.
- LLM-generated indicator calculations or inferred bid/ask/spread.

---

## MCP Tools Safe To Use

Safe means the tool is either read-only or has clear operational behavior and is
not known to return fake numerical truth for HTS decisions.

| Tool group | Tools |
| --- | --- |
| Health/discovery | `tv_health_check`, `tv_discover`, `tv_ui_state`, `tv_launch` |
| Chart state/control | `chart_get_state`, `chart_set_symbol`, `chart_set_timeframe`, `chart_set_chart_type`, `chart_wait_ready` |
| Data | `data_get_ohlcv` |
| Quote partial | `quote_get` for last/OHLC/volume only; not BBO when unavailable |
| Symbol | `symbol_search`, `symbol_info` |
| Capture | `capture_screenshot` |
| Panes/tabs/layout | pane and tab listing/focus/switch helpers, with UI-state caveats |
| UI read helpers | `ui_find_element`, `ui_evaluate` when used for diagnostics |
| Pine read-only | `pine_get_source`, `pine_get_errors`, `pine_list_scripts`, `pine_analyze`, `pine_check` |
| Replay read/control | `replay_status`, `replay_start`, `replay_step`, `replay_stop`, `replay_autoplay` for practice workflows |
| Indicator control | `indicator_set_inputs`, `indicator_toggle_visibility` for applied indicators with verification |
| HTS helpers | `continuous_contract_context` as analysis-only metadata |

---

## MCP Tools Requiring Work

| Tool | Required work |
| --- | --- |
| `data_get_indicator` | Fix locale parser, add `source`, `raw_value`, `reliable_for_trading_logic`, `warning`, and stable name mapping. |
| `data_get_study_values` | Same source/reliability flags; avoid presenting display text as raw study values. |
| New `data_get_indicator_history` | Expose internal `study.data().valueAt(index)` with `metaInfo().plots` mapping. |
| `quote_get` | Add `bid_available`, `ask_available`, `bid_source`, `ask_source`, `warning`; distinguish absent keys from real zero. |
| `data_get_strategy_results` | Return explicit `no_strategy_loaded` / `strategy_data_unavailable`, not empty success. |
| `data_get_trades` | Align success/error contract and verify against loaded strategy. |
| `data_get_equity` | Same no-strategy/unavailable handling and live verification. |
| `chart_manage_indicator` | Do not return full success when add returns empty `entityId` or study count does not change. |
| `pine_compile` | Make UI action explicit and verify add/update/save side effects. |
| `pine_smart_compile` | Distinguish compile OK, save OK, add-to-chart OK, and update-on-chart OK. |
| `pine_get_console` | Mark as noisy DOM scrape or replace with structured source when possible. |
| `replay_trade` | Do not imply confirmed trade state when position/PnL remain null. |
| HTS composite `market_summary` | Add source/quality flags or rebuild in external HTS from trusted inputs. |

---

## Data To Calculate In HTS Go

These should be deterministic HTS calculations from trusted OHLCV/orderbook
inputs, not LLM calculations:

- RSI.
- ADX / DI+ / DI-.
- EMA, SMA, KAMA, slope, cross state.
- ATR and ATR percent.
- Realized volatility and volatility regime.
- Volume regime and volume vs average.
- Market phase classification.
- Trend direction and trend strength.
- Distance to levels.
- Risk/reward.
- Position sizing.
- Diff summaries for re-evaluation.
- Trade review attribution and rule checks.
- Scanner scores/ranking for multi-symbol lists.

LLM receives compact summaries, features, quality flags, and warnings. It does
not receive raw candles by default.

---

## Data To Take From TradingView

TradingView MCP is appropriate for:

- analysis symbol and chart state;
- timeframe and chart type;
- TradingView OHLCV bars;
- current last/OHLC/volume fields when present;
- loaded indicator metadata and input settings;
- applied study internal history after `data_get_indicator_history` is
  implemented and verified;
- Pine source only when visible/authorized for the current user;
- prepared Pine `HTS_JSON` output from tables/labels when schema/version/script
  identity are verified;
- drawings, levels, labels, boxes, and screenshots for visual context;
- replay state for training/practice workflows;
- account/plan/indicator-limit diagnostics where available.

TradingView should not be the source for executable MOEX BBO when bid/ask keys
are absent.

---

## Data To Take From Tinkoff

Tinkoff should be the source for execution-facing data:

- concrete tradable futures instrument;
- `instrument_uid`, FIGI, ticker, class code;
- expiration and last trade date;
- lot and min price increment;
- trading status and API trading availability;
- futures margin;
- orderbook;
- bid/ask/spread;
- account size, positions, and orders when the external HTS layer needs them.

The required boundary is documented in
`docs/dev/INSTRUMENT_RESOLVER_CONTRACT.md`.

---

## Blocking Issues

Blocking for full HTS numerical automation:

1. Indicator values are unsafe until locale parser/source flags are fixed or
   replaced by internal study history/local Go calculations.
2. Strategy metrics/trades/equity are not a reliable contract yet.
3. MOEX bid/ask/spread are unavailable from tested TradingView state and must
   come from Tinkoff orderbook.
4. Execution symbol resolution is only documented, not implemented in this repo.
5. `chart_manage_indicator` can report success when add did not create a study.
6. Pine compile/save/open workflows can mutate user scripts and need safer
   explicit modes.

These issues block `GO` but do not block `GO WITH LIMITATIONS`.

---

## Non-Blocking Issues

These should be fixed, but they do not stop analysis-only HTS integration:

- `symbol_search` endpoint should have page-context fallback if direct API
  changes again.
- Screenshot CLI/MCP live regression should be run after filename fix.
- Replay trade position/PnL confirmation is weak.
- `pine_get_console` is noisy and should be diagnostic-only.
- Strategy input mutation is not live-confirmed with a loaded strategy.
- Existing docs still need cleanup in places that describe indicator values as
  canvas/Y coordinates; current evidence points to Data Window locale parsing.
- SQLite resolver schema is documented but not parse-tested in this environment
  because `sqlite3` is not installed.

---

## Next Actions

Priority 1:

1. Fix locale-aware numeric parser for TradingView display values.
2. Add source/reliability fields to indicator tools.
3. Implement `data_get_indicator_history` using `study.data().valueAt(index)`
   and `metaInfo().plots` slot mapping.
4. Update `quote_get` with bid/ask availability/source flags.
5. Fix strategy tools to return explicit no-strategy/unavailable statuses.

Priority 2:

1. Implement external HTS `InstrumentResolver`.
2. Integrate Tinkoff orderbook for executable bid/ask/spread.
3. Integrate Tinkoff futures expiration, margin, status, lot, min price
   increment.
4. Build HTS Go feature calculators for RSI/ADX/EMA/KAMA/ATR/phase/trend.
5. Build compact LLM payload generator from trusted fields only.

Priority 3:

1. Add safe Pine compile modes that separate check, save, add-to-chart, and
   update-on-chart.
2. Add live smoke tests for loaded strategy results/trades/equity.
3. Add page-context fallback for `symbol_search`.
4. Add final docs cleanup so old status files no longer overstate canvas/Y
   indicator claims.
5. Add resolver schema parse test in the external HTS repo once SQLite is
   available.

---

## Decision

```text
GO WITH LIMITATIONS
```

Rationale:

- `GO` is not justified because current indicator values, strategy metrics, and
  executable bid/ask are not reliable enough for autonomous numerical decisions.
- `NO-GO` is too strict because TradingView MCP is already useful and tested for
  chart context, OHLCV, screenshots, symbol search, Pine source access, replay
  practice, and operator-facing analysis.
- `GO WITH LIMITATIONS` matches the evidence: use this repo as TradingView
  data/visual/context collector, while HTS Go calculates critical features and
  Tinkoff provides execution-grade instrument and orderbook data.
