# Strategy Equity Extraction Implementation Smoke

Date: 2026-04-27
OS: Windows
Chart: `RUS:NG1!`, `1D`

## Scope

`data_get_equity` now prioritizes explicit Pine runtime output:

```pine
plot(strategy.equity, "Strategy Equity", display=display.data_window)
```

When that plot exists on the active strategy, the tool reads:

```js
model.strategySources()[0].data().fullRangeIterator()
```

The plot is found through `metaInfo().plots` and `metaInfo().styles`; returned
points have:

- `index`
- `time` in milliseconds
- `equity`

The successful response is marked:

```json
{
  "source": "tradingview_strategy_plot",
  "coverage": "loaded_chart_bars",
  "reliability": "reliable_pine_runtime_value_unstable_internal_path",
  "reliableForTradingLogic": true
}
```

Loaded bars are explicitly not claimed to be the full historical backtest.

## Missing Plot Behavior

If a strategy is loaded but the explicit `Strategy Equity` plot is absent,
`data_get_equity` returns:

```json
{
  "success": false,
  "status": "needs_equity_plot",
  "suggested_pine_line": "plot(strategy.equity, \"Strategy Equity\", display=display.data_window)"
}
```

The response also includes a `derived_fallback` descriptor with:

- `source: derived_from_ohlcv_and_trades`
- `reliableForTradingLogic: false`
- requirements and limitations

If the backtesting report exposes `trades[].cumulativeProfit`, the descriptor
may include trade-exit points, but they are marked as `trade_exit_points_only`,
not bar-by-bar equity.

## Live Smoke

Files:

- `mcp-equity-smoke.output.jsonl`
- `mcp-equity-smoke.summary.json`
- `mcp-equity-smoke.stderr.txt`

Current chart has only ordinary indicators:

```text
Vvzmzg: Помошник RSI - True ADX
cfyrsD: Volume
```

No strategy was loaded, so the live `data_get_equity` result is:

```json
{
  "success": false,
  "status": "no_strategy_loaded"
}
```

The strategy-equity-plot path was not re-smoked with a disposable live strategy
in this run to avoid mutating the current chart/Pine state. The previously
confirmed live strategy plot evidence remains in `research/strategy-equity-full/`.

## Tests

```text
go test ./internal/tools/data
go test ./...
go vet ./...
```

All passed on 2026-04-27.
