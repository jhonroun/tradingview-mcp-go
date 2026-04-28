# Strategy Backtesting API Implementation Smoke

Date: 2026-04-27
OS: Windows
Chart: `RUS:NG1!`, `1D`

## Scope

Implemented strategy tools now use:

- `model.strategySources()`
- `model.activeStrategySource()`
- `await window.TradingViewApi.backtestingStrategyApi()`

The old detector based on `dataSources() + performance` is not used by the
strategy extraction helper.

## Files

- `mcp-strategy-tools.output.jsonl` - raw MCP initialize, tools/list,
  chart_get_state, and strategy tool calls.
- `mcp-strategy-tools.summary.json` - compact parsed result.
- `mcp-tools-names.json` - sorted tool names from tools/list.
- `mcp-strategy-tools.stderr.txt` - stderr banner only.

## Live Result

`tools/list` returned 84 tools, including the new `data_get_orders`.

Current chart studies:

```text
Vvzmzg: Помошник RSI - True ADX
cfyrsD: Volume
```

No strategy was loaded during this smoke:

```json
{
  "success": false,
  "status": "no_strategy_loaded",
  "source": "tradingview_backtesting_api",
  "strategy_loaded": false,
  "strategy_source_count": 0
}
```

The same `no_strategy_loaded` status was returned by:

- `data_get_strategy_results`
- `data_get_trades`
- `data_get_orders`
- `data_get_equity`

This confirms ordinary indicators were not classified as strategies.

## Strategy-Loaded Verification

This implementation was not re-smoked with a live loaded test strategy in this
run to avoid mutating the current chart/Pine state. The confirmed report shape
from the earlier live strategy research is in `research/strategy-live-test/`:

- `report.trades`
- `report.filledOrders`
- `report.performance`
- `report.settings`
- `report.currency`

The new helper maps those fields through `backtestingStrategyApi()` and returns
`status: "ok"` when the report shape is available.

## Tests

```text
go test ./internal/tools/data ./internal/mcp
go test ./...
go vet ./...
```

All passed on 2026-04-27.
