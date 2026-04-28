# Strategy Live Test: TradingView Backtesting API Without CSV

Date: 2026-04-27
Chart: `RUS:NG1!`, `1D`
Path tested: `window.TradingViewApi.backtestingStrategyApi()`

## Result

Extraction is possible through TradingView internals without CSV export.

The loaded test strategy was visible in both places:

- chart studies: `MCP Test SMA Strategy`, entity id `XTcDfv`;
- backtesting API: `allStrategies[0]`, `activeStrategy`, `isStrategyEmpty=false`.

`activeStrategyReportData` was non-null and contained:

```text
currency
settings
buyHold
buyHoldPercent
filledOrders
performance
trades
firstTradeIndex
activeOrders
runupDrawdownPeriods
hasRunupData
hasDrawdownData
```

Confirmed programmatic data:

- trades: `report.trades`, 43 rows;
- entry/exit: `report.trades[].entry` / `report.trades[].exit`;
- PnL: `report.trades[].profit` and `report.trades[].cumulativeProfit`;
- orders: `report.filledOrders`, 86 rows;
- metrics: `report.performance.all` and summary fields under
  `report.performance`;
- equity: no full bar-by-bar equity array was observed, but a close-to-close
  equity series can be reconstructed from
  `report.trades[].cumulativeProfit + report.performance.initialCapital`.

## Files

- `00-pine-editor-backup.json` - full Pine Editor source backup before mutation.
- `00-pine-editor-backup.pine` - extracted backup used for restore.
- `01-before-state.json` - chart state, data sources, no active strategy.
- `02-limit-handling.json` - Basic plan limit evidence and indicator removal.
- `03-after-add-strategy.json` - backtesting API after strategy was added.
- `04-model-strategy-api.json` - `model.strategySources()` and active strategy
  source snapshot.
- `05-backtesting-report.json` - `activeStrategyReportData` preview.
- `06-report-key-search.json` - recursive key search for trades/orders/equity.
- `07-normalized-result.json` - normalized extraction prototype.
- `08-cleanup.json`, `08b-pine-restore-check.json`,
  `09-restore-custom-indicator-attempt.json`, `10-restore-volume.json` -
  cleanup and restore evidence.

## Limit Handling

The first add attempt compiled and reported "Added to chart", but TradingView
also showed the Basic subscription limit:

```text
You applied 2 indicators - maximum available for your subscription.
Current subscription: Basic, 2.
```

The limit dialog counted `HSfT` / `Help system for trade` and `RSI-ADX`.
`Volume` did not appear in that limit dialog.

To make room for the test strategy, `Help system for trade` was removed:

```json
{
  "id": "bzUmEv",
  "name": "Help system for trade"
}
```

This follows the task instruction: if the indicator limit blocks adding a
strategy, remove one indicator and record which one.

## Cleanup

After the research data was captured:

- test strategy `XTcDfv` was removed from the chart;
- Pine Editor source was restored exactly from backup:
  `e675eb546b9f96fb79c1b7cd179d8908e7a77c7fcb4ca798124166042d350765`;
- `Volume` was re-added, with a new entity id `cfyrsD`;
- `Help system for trade` could not be restored through `createStudy`:
  TradingView did not throw an exception, but the study list did not change.

Final chart studies after cleanup:

```text
Помошник RSI - True ADX
Volume
```

## Access Paths

Strategy identity:

```js
const chart = window.TradingViewApi._activeChartWidgetWV.value();
const model = chart._chartWidget.model().model();

const strategies = model.strategySources();
const activeStrategy = model.activeStrategySource().value();
```

Backtesting report:

```js
const bt = await window.TradingViewApi.backtestingStrategyApi();

const activeStrategy = bt.activeStrategy.value();
const report = (
  bt.activeStrategyReportData ||
  bt._activeStrategyReportData ||
  bt._reportData
).value();
```

Data extraction:

```js
const metrics = report.performance.all;
const trades = report.trades;
const orders = report.filledOrders;

const equity = report.trades.map((trade) => ({
  time: trade.exit.time,
  equity: report.performance.initialCapital + trade.cumulativeProfit.value,
  cumulative_pnl: trade.cumulativeProfit.value
}));
```

## Reliability

| Method | Reliability | Result |
| --- | --- | --- |
| `backtestingStrategyApi()` watched values | reliable live, undocumented | Strategy identity and report loaded correctly. |
| `activeStrategyReportData.value().trades` | reliable live, undocumented | Contains entry/exit/time/PnL for 43 trades. |
| `activeStrategyReportData.value().filledOrders` | reliable live, undocumented | Contains 86 filled orders. |
| `activeStrategyReportData.value().performance` | reliable live, undocumented | Contains summary metrics. |
| Equity from `trades[].cumulativeProfit` | reliable for close-to-close reconstruction | Not full bar-by-bar equity. |
| Strategy Tester DOM | ui-based fallback only | Not needed for extraction. |
| CSV export | not used | Prohibited by task. |

## MCP Tool Feasibility

`data_get_strategy_results`: implementable now from:

```js
report.performance.all
report.performance
```

`data_get_trades`: implementable now from:

```js
report.trades
```

`data_get_equity`: implementable with limitations:

```js
report.trades[].cumulativeProfit + report.performance.initialCapital
```

This gives trade-exit equity points. Full per-bar equity was not found in the
report object. For full equity, use a prepared Pine output such as
`plot(strategy.equity, "HTS_equity")` or continue deeper internal research.
