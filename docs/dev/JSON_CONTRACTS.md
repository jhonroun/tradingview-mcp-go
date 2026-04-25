# JSON Contracts — HTS integration

Response structures for the tools consumed by the HTS layer.
These are the authoritative field names and types Go currently returns.

> Status: **draft** — verify against live TradingView before marking stable.

---

## Error envelope (all tools)

Every tool returns one of these two shapes. The outer JSON-RPC wrapper is stripped by the MCP client before the HTS layer sees this.

```json
{ "success": false, "error": "<human-readable message>" }
{ "success": true,  ...tool-specific fields... }
```

### Retryable vs permanent errors

| Condition | `error` contains | Retryable |
| --- | --- | --- |
| CDP not connected | `"CDP"` or `"connect"` | yes — wait and retry |
| TradingView tab not found | `"no TradingView"` | yes — reopen tab |
| JS evaluation timeout | `"timeout"` | yes |
| Unknown tool name | `"unknown tool"` | no |
| Invalid argument type | `"unmarshal"` or `"invalid"` | no |

---

## chart_get_state

```json
{
  "success": true,
  "symbol": "BINANCE:BTCUSDT",
  "exchange": "BINANCE",
  "ticker": "BTCUSDT",
  "timeframe": "60",
  "type": "Candles",
  "indicators": [
    {
      "id": "Study_RSI_0",
      "name": "Relative Strength Index",
      "inputs": { "length": 14 }
    }
  ],
  "pane_count": 2
}
```

Notes:

- `timeframe`: string — `"1"` `"5"` `"15"` `"60"` `"240"` `"D"` `"W"` `"M"`
- `indicators`: may be empty array `[]`
- `inputs`: object shape varies per indicator; keys are Pine parameter names

---

## quote_get

```json
{
  "success": true,
  "symbol": "BINANCE:BTCUSDT",
  "last":   67400.0,
  "open":   66800.0,
  "high":   67900.0,
  "low":    66500.0,
  "close":  67400.0,
  "volume": 12345.67,
  "bid":    67398.0,
  "ask":    67402.0,
  "change": 600.0,
  "change_pct": 0.90
}
```

Notes:

- `bid`/`ask`: may be `0` or absent for non-orderbook symbols (indices, crypto on some feeds)
- All price fields are `float64`; never null — use `0` as sentinel
- `change_pct`: percentage as float (`0.90` = 0.90%, not 0.0090)

---

## data_get_ohlcv

```json
{
  "success": true,
  "bar_count": 100,
  "total_available": 5000,
  "source": "direct_bars",
  "bars": [
    {
      "time":   1713916800,
      "open":   66800.0,
      "high":   67900.0,
      "low":    66500.0,
      "close":  67400.0,
      "volume": 12345.67
    }
  ]
}
```

With `summary=true`:

```json
{
  "success": true,
  "bar_count": 100,
  "period": { "from": 1713916800, "to": 1714003200 },
  "open":   66800.0,
  "close":  67400.0,
  "high":   67900.0,
  "low":    66500.0,
  "range":  1400.0,
  "change": 600.0,
  "change_pct": "0.90%",
  "avg_volume": 11000.0,
  "last_5_bars": [ ...same bar shape... ]
}
```

Notes:

- `bars[0]` = oldest bar; `bars[bar_count-1]` = most recent bar
- `time`: Unix timestamp seconds (UTC)
- `source`: `"direct_bars"` | `"study_bars"` | `"fallback"`

---

## data_get_study_values

```json
{
  "success": true,
  "study_count": 2,
  "studies": [
    {
      "name": "RSI",
      "entity_id": "Study_RSI_0",
      "plot_count": 1,
      "plots": [
        {
          "name": "RSI",
          "values": [55.3, 54.1, 56.8],
          "current": 55.3
        }
      ]
    }
  ]
}
```

Notes:

- `values`: newest-first — `values[0]` = current bar
- `current`: shorthand alias for `values[0]`; always present if `values` non-empty
- `studies` may be empty array when no indicators are on chart
- Multi-output indicators (e.g. Bollinger Bands) have multiple entries in `plots`

---

## data_get_indicator

Single indicator by `entity_id`:

```json
{
  "success": true,
  "entity_id": "Study_RSI_0",
  "name": "Relative Strength Index",
  "inputs": { "length": 14, "source": "close" },
  "plots": [
    {
      "name": "RSI",
      "values": [55.3, 54.1, 56.8],
      "current": 55.3
    }
  ]
}
```

Notes:

- `inputs` fields with values >500 chars are omitted (large Pine source strings)
- `inputs` fields with values >200 chars are truncated with `"...(truncated)"`

---

## symbol_info

```json
{
  "success": true,
  "symbol":      "BTCUSDT",
  "exchange":    "BINANCE",
  "description": "Bitcoin / TetherUS",
  "type":        "crypto",
  "currency":    "USDT",
  "timezone":    "Etc/UTC",
  "session":     "24x7",
  "minmov":      1,
  "pricescale":  100
}
```

Notes:

- `type`: `"stock"` | `"crypto"` | `"forex"` | `"futures"` | `"index"` | `"fund"` | `"dr"` | `"bond"`
- `session`: `"24x7"` for crypto; exchange hours string for equities

---

## symbol_search

```json
{
  "success": true,
  "count": 3,
  "results": [
    {
      "symbol":      "BTCUSDT",
      "exchange":    "BINANCE",
      "description": "Bitcoin / TetherUS",
      "type":        "crypto"
    }
  ]
}
```

Notes:

- `count` ≤ 15 (hard limit in TradingView search API)
- All four fields always present; `description` may be empty string

---

## HTS consumption pattern

```
TradingView computes                   →  indicator values are ready
tradingview-mcp-go fetches             →  data_get_study_values / quote_get / chart_get_state
HTS normalises                         →  maps raw values to typed domain structs
DeepSeek / LLM explains               →  receives normalised context, generates analysis
```

Minimum viable context object for one LLM call (future `chart_context_for_llm`):

```json
{
  "symbol":    "BINANCE:BTCUSDT",
  "timeframe": "60",
  "bar": { "time": 1714003200, "open": 66800, "high": 67900, "low": 66500, "close": 67400, "volume": 12345 },
  "change_pct": 0.90,
  "indicators": [
    { "name": "RSI",  "current": 55.3 },
    { "name": "MACD", "current": 120.5, "signal": 115.2, "histogram": 5.3 }
  ]
}
```
