# TradingView Strategy Equity Curve Research

Date: 2026-04-27
Symbol: `RUS:NG1!`
Timeframe: `1D`
Strategy used: `MCP Test SMA Strategy`

## Result

Full bar-by-bar strategy equity was not found as a native series in
`backtestingStrategyApi()` report objects. The report exposes trades, orders,
summary metrics, buy-and-hold samples, and runup/drawdown periods, but not a
per-candle strategy equity array.

A working machine-readable path was confirmed if the Pine strategy explicitly
plots `strategy.equity`:

```pine
plot(strategy.equity, "Strategy Equity", display=display.data_window)
```

With that plot present, the equity series is available through the same
internal study data path already used for indicator values:

```js
const chart = window.TradingViewApi._activeChartWidgetWV.value();
const model = chart._chartWidget.model().model();
const src = model.strategySources()[0];
const data = src.data();

const plots = src.metaInfo().plots;
const styles = src.metaInfo().styles || {};
const equityPlot = plots.findIndex((p) => {
  const style = styles[p.id] || {};
  return style.title === "Strategy Equity" || p.id === "plot_2";
});
if (equityPlot < 0) throw new Error("Strategy Equity plot not found");

const equityOffset = equityPlot + 1; // row[0] is timestamp in seconds
const rows = [];

for (let it = data.fullRangeIterator(), r = it.next(); !r.done; r = it.next()) {
  const row = r.value;
  rows.push({
    index: row.index,
    time: row.value[0] * 1000,
    equity: row.value[equityOffset],
  });
}

return rows;
```

Confirmed sample shape:

```json
[
  { "index": -1000101, "time": 1727366400000, "equity": 1000565.3 },
  { "index": -1000100, "time": 1727452800000, "equity": 1000568.5 },
  { "index": 298, "time": 1777006800000, "equity": 1000224 }
]
```

The final equity `1000224` matches `initialCapital + netProfit`
(`1000000 + 224`).

## Checked Paths

Native report path:

```js
const bt = await window.TradingViewApi.backtestingStrategyApi();
const report = (
  bt.activeStrategyReportData ||
  bt._activeStrategyReportData ||
  bt._reportData
).value();
```

Observed top-level report keys:

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

Observed lengths:

```text
report.trades: 43
report.filledOrders: 86
report.buyHold: 44
report.buyHoldPercent: 44
report.runupDrawdownPeriods: 5
```

`buyHold` and `buyHoldPercent` are not strategy equity. They are buy-and-hold
values and have no per-bar timestamps. Their length is close to trade count,
not chart bar count.

Deep backtesting path:

```js
bt._deepBacktestingManager._reportDataDeepBacktesting
```

Observed value: `null`. Deep backtesting was not active.

Model strategy source paths:

```js
const src = model.strategySources()[0];
src.reportData();
src.ordersData();
src.barsIndexes();
src.data();
```

Observed:

```text
src.reportData(): compact report copy, no full equity series
src.ordersData(): 86 order records
src.barsIndexes(): 86 order bar indexes
src.data(): plotted Pine series
```

Before adding `plot(strategy.equity)`, `src.data().valueAt(index)` returned:

```js
[time, fastSMA, slowSMA]
```

After adding `plot(strategy.equity)`, it returned:

```js
[time, fastSMA, slowSMA, equity]
```

## Reliability

`report.trades[].cumulativeProfit`: reliable, but trade-exit only. Not a
bar-by-bar equity curve.

`src.data().fullRangeIterator()` with explicit `plot(strategy.equity)`:
reliable as TradingView Pine runtime output for loaded bars, but the access path
is undocumented internal TradingView state and should be marked `unstable`.

Fallback reconstruction from OHLCV and trades: `derived`. It can produce
bar-by-bar equity, but depends on correct OHLCV coverage, trade direction,
quantity, point value, commission/slippage settings, and fill timing.

## Coverage Limitation

In this live session, the plotted strategy source had `data.size() == 400`.
The report date range covered 2020-02-01 to 2026-04-25, while the plotted
series covered the loaded chart data window from 2024-09-26 to 2026-04-24.

So the confirmed extraction path returns all bars currently present in
`src.data()`, not automatically every historical backtest bar if TradingView has
not loaded those bars into the chart model.

## Derived Fallback

If no explicit `strategy.equity` plot is present, reconstruct equity from OHLCV
and `report.trades`:

```js
function reconstructEquity({ bars, trades, initialCapital, pointValue }) {
  const result = [];

  for (const bar of bars) {
    const t = bar.time;
    let realized = 0;
    let openPnL = 0;

    for (const tr of trades) {
      const entryTime = tr.entry.time;
      const exitTime = tr.exit.time;
      const qty = tr.quantity || 1;

      if (exitTime <= t) {
        realized = tr.cumulativeProfit.value;
        continue;
      }

      if (entryTime <= t && t < exitTime) {
        const side = tr.entry.type === "se" ? -1 : 1;
        openPnL += side * qty * pointValue * (bar.close - tr.entry.price);
      }
    }

    result.push({
      time: t,
      equity: initialCapital + realized + openPnL,
      source: "derived_from_ohlcv_and_trades",
    });
  }

  return result;
}
```

Production reconstruction must handle short trades, pyramiding, partial fills,
commissions, slippage, order execution timing, and contract `pointValue`.

## Cleanup

No indicator was removed for this test.

The disposable strategy `9MB0zA` was removed after research. Final chart studies:

```text
Vvzmzg: Помошник RSI - True ADX
cfyrsD: Volume
```

Pine Editor source was restored from
`research/strategy-live-test/00-pine-editor-backup.pine`.

Restored SHA256:

```text
e675eb546b9f96fb79c1b7cd179d8908e7a77c7fcb4ca798124166042d350765
```
