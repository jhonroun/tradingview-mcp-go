# Status

```text
tradingview-mcp-go = working TradingView MCP/CLI base for HTS integration
status: GO WITH LIMITATIONS
final audit: docs/dev/FINAL_AUDIT_REPORT.md
```

Confirmed:

```text
MCP registry: 85 tools registered in the current stabilization branch
CLI: implemented
CDP connection: works against TradingView Desktop
chart_get_state: live-tested
quote_get: live-tested for last/OHLC/volume; bid/ask availability is explicit
data_get_ohlcv: live-tested
symbol_search: live-tested; empty responses include status/reason
data_get_indicator: live-tested through TradingView study model
data_get_indicator_history: live-tested through TradingView study model
capture_screenshot: live-tested; .png.png filename bug fixed/unit-tested
pine_get_source: live_tested
strategy tools: partial; current chart returns structured no_strategy_loaded
replay_status: partial live test; replay trade workflow unverified
MCP stdin reader: Reader.ReadBytes + 16 MB guard
```

Important limitations:

```text
Indicator values from `data_get_indicator`, `data_get_indicator_history`, and
`data_get_study_values` use TradingView's internal study model and are marked:

source: tradingview_study_model
reliability: reliable_pine_runtime_value_unstable_internal_path
reliableForTradingLogic: true

UI/Data Window display string parsing remains a fallback-only path and is
marked unreliable for trading logic.

MOEX futures bid/ask were unavailable in tested TradingView quote state.
`quote_get` now marks this with bidAskAvailable=false and sourceLimitation.

Strategy metrics/trades/orders/equity use TradingView backtesting internals but
remain partial until a loaded-strategy live smoke is repeated on the current
branch. No-strategy states are structured and no longer silent success.

Replay status is readable, but replay trade workflow remains unverified in the
current stabilization pass.
```

Decision:

```text
GO WITH LIMITATIONS
```

Use this repository as TradingView chart/context/visual/Pine collection layer.
Calculate critical features in HTS Go and use Tinkoff for execution-grade
instrument identity, orderbook, expiration, margin, and trading status.
