# HTS Market Summary Contract

Status: draft, documentation only.

This contract describes the object passed from a future HTS MCP layer to an LLM
after collecting data from TradingView MCP, HTS Go calculations, prepared
PineScript output, and Tinkoff market data.

Boundary: `tradingview-mcp-go` may produce TradingView-sourced inputs and can
document this contract, but Tinkoff integration, execution symbol resolution,
risk sizing, and trading decisions belong outside this repository.

For the compact payloads sent from HTS to an LLM after this market summary is
normalized, see [LLM_MARKET_CONTEXT_CONTRACT.md](LLM_MARKET_CONTEXT_CONTRACT.md).
For the proposed analysis-symbol to execution-instrument resolver, see
[INSTRUMENT_RESOLVER_CONTRACT.md](INSTRUMENT_RESOLVER_CONTRACT.md).

---

## Source Classes

Every numeric or decision-support field must carry a source marker directly or
through `source_ref`.

| Source class | Meaning | Can be used as LLM truth |
| --- | --- | --- |
| `tradingview_direct` | TradingView chart state, quote fields, OHLCV bars, or internal study model values read through MCP. | Yes, when `status="ok"` and not stale. |
| `hts_go_derived` | Deterministic HTS calculation from trusted inputs, for example ATR, EMA slope, phase, trend, volatility regime. | Yes as HTS-derived assessment, not as raw market data. |
| `pine_hts_json` | Prepared PineScript emits strict machine-readable `HTS_JSON` through a table/label/other structured output. | Yes, if schema/version/script hash are verified. |
| `tinkoff_marketdata` | Tinkoff Invest instruments, last price, orderbook, bid/ask, spread, lot, min price increment. | Yes for executable MOEX bid/ask/spread when fresh. |
| `unreliable` | Canvas coordinates, pixel reads, loosely parsed DOM/Data Window text, zero sentinel values, empty strategy metrics, or unknown source. | No. |

Recommended field status values:

- `ok`: value is present, fresh, and source is trusted.
- `partial`: value is present but incomplete.
- `unavailable`: source does not provide the value.
- `stale`: value is older than the allowed freshness window.
- `unreliable`: value must not be treated as factual.
- `error`: source failed and the field is not usable.

Recommended reliability values:

- `trusted`
- `usable_with_warning`
- `unavailable`
- `stale`
- `unreliable`

---

## Top-Level Contract

Required top-level fields:

- `analysis_symbol`
- `execution_symbol`
- `timeframe`
- `current_price`
- `ohlcv_summary`
- `phase`
- `trend`
- `volatility`
- `levels`
- `signals`
- `risk_context`
- `data_quality`
- `warnings`
- `recommended_recheck_at`
- `source_trace`

Additional recommended fields:

- `contract_version`
- `success`
- `generated_at`

Field rules:

- `analysis_symbol` is the chart symbol used for analysis. It may be a
  continuous TradingView futures contract such as `RUS:NG1!`.
- `execution_symbol` is the real tradable instrument selected by HTS/Tinkoff.
  It must not silently copy a continuous TradingView contract.
- `current_price` can come from TradingView or Tinkoff. The selected source must
  be explicit.
- `risk_context.bid`, `risk_context.ask`, and `risk_context.spread` should come
  from Tinkoff orderbook for MOEX execution decisions. TradingView `bid=0` or
  absent bid/ask must be marked unavailable or unreliable.
- `phase`, `trend`, `volatility`, `levels`, and `signals` are HTS assessment
  fields. They must expose their evidence through `source_ref` and must not hide
  source limitations.

---

## Proposed Go Structs

These structs are proposed for the external HTS MCP layer. They are not
implemented in `tradingview-mcp-go`.

```go
package htscontract

import "time"

type SourceClass string

const (
	SourceTradingViewDirect SourceClass = "tradingview_direct"
	SourceHTSGoDerived     SourceClass = "hts_go_derived"
	SourcePineHTSJSON      SourceClass = "pine_hts_json"
	SourceTinkoffMarket    SourceClass = "tinkoff_marketdata"
	SourceUnreliable       SourceClass = "unreliable"
)

type FieldStatus string

const (
	StatusOK          FieldStatus = "ok"
	StatusPartial     FieldStatus = "partial"
	StatusUnavailable FieldStatus = "unavailable"
	StatusStale       FieldStatus = "stale"
	StatusUnreliable  FieldStatus = "unreliable"
	StatusError       FieldStatus = "error"
)

type Reliability string

const (
	ReliabilityTrusted           Reliability = "trusted"
	ReliabilityUsableWithWarning Reliability = "usable_with_warning"
	ReliabilityUnavailable       Reliability = "unavailable"
	ReliabilityStale             Reliability = "stale"
	ReliabilityUnreliable        Reliability = "unreliable"
)

type MarketSummary struct {
	ContractVersion      string          `json:"contract_version"`
	Success              bool            `json:"success"`
	GeneratedAt          time.Time       `json:"generated_at"`
	AnalysisSymbol       SymbolRef       `json:"analysis_symbol"`
	ExecutionSymbol      ExecutionSymbol `json:"execution_symbol"`
	Timeframe            TimeframeRef    `json:"timeframe"`
	CurrentPrice         PriceSnapshot   `json:"current_price"`
	OhlcvSummary         OhlcvSummary    `json:"ohlcv_summary"`
	Phase                PhaseSummary    `json:"phase"`
	Trend                TrendSummary    `json:"trend"`
	Volatility           Volatility      `json:"volatility"`
	Levels               []Level         `json:"levels"`
	Signals              []Signal        `json:"signals"`
	RiskContext          RiskContext     `json:"risk_context"`
	DataQuality          DataQuality     `json:"data_quality"`
	Warnings             []Warning       `json:"warnings"`
	RecommendedRecheckAt time.Time       `json:"recommended_recheck_at"`
	SourceTrace          []SourceTrace   `json:"source_trace"`
}

type SymbolRef struct {
	Symbol       string      `json:"symbol"`
	Exchange     string      `json:"exchange,omitempty"`
	Ticker       string      `json:"ticker,omitempty"`
	Type         string      `json:"type,omitempty"`
	IsContinuous bool        `json:"is_continuous,omitempty"`
	SourceRef    string      `json:"source_ref"`
	Status       FieldStatus `json:"status"`
	Reliability  Reliability `json:"reliability"`
}

type ExecutionSymbol struct {
	Symbol            string      `json:"symbol,omitempty"`
	Ticker            string      `json:"ticker,omitempty"`
	ClassCode         string      `json:"class_code,omitempty"`
	FIGI              string      `json:"figi,omitempty"`
	InstrumentUID     string      `json:"instrument_uid,omitempty"`
	Lot               int         `json:"lot,omitempty"`
	MinPriceIncrement float64     `json:"min_price_increment,omitempty"`
	MappedFrom         string      `json:"mapped_from,omitempty"`
	SourceRef          string      `json:"source_ref"`
	Status             FieldStatus `json:"status"`
	Reliability        Reliability `json:"reliability"`
}

type TimeframeRef struct {
	Value       string      `json:"value"`
	Seconds     int64       `json:"seconds,omitempty"`
	SourceRef   string      `json:"source_ref"`
	Status      FieldStatus `json:"status"`
	Reliability Reliability `json:"reliability"`
}

type PriceSnapshot struct {
	Last        float64     `json:"last,omitempty"`
	Bid         *float64    `json:"bid,omitempty"`
	Ask         *float64    `json:"ask,omitempty"`
	Mid         *float64    `json:"mid,omitempty"`
	Currency    string      `json:"currency,omitempty"`
	AsOf        *time.Time  `json:"as_of,omitempty"`
	SourceRef   string      `json:"source_ref"`
	Status      FieldStatus `json:"status"`
	Reliability Reliability `json:"reliability"`
}

type OhlcvBar struct {
	Time   time.Time `json:"time"`
	Open   float64   `json:"open"`
	High   float64   `json:"high"`
	Low    float64   `json:"low"`
	Close  float64   `json:"close"`
	Volume float64   `json:"volume"`
}

type OhlcvSummary struct {
	BarCount     int         `json:"bar_count"`
	PeriodFrom   *time.Time  `json:"period_from,omitempty"`
	PeriodTo     *time.Time  `json:"period_to,omitempty"`
	LastBar      OhlcvBar    `json:"last_bar"`
	Change       float64     `json:"change,omitempty"`
	ChangePct    float64     `json:"change_pct,omitempty"`
	AverageVolume float64    `json:"average_volume,omitempty"`
	VolumeVsAvg  float64     `json:"volume_vs_avg,omitempty"`
	SourceRef    string      `json:"source_ref"`
	Status       FieldStatus `json:"status"`
	Reliability  Reliability `json:"reliability"`
}

type PhaseSummary struct {
	Value       string      `json:"value"`
	Confidence  float64     `json:"confidence,omitempty"`
	EvidenceRefs []string    `json:"evidence_refs,omitempty"`
	SourceRef   string      `json:"source_ref"`
	Status      FieldStatus `json:"status"`
	Reliability Reliability `json:"reliability"`
}

type TrendSummary struct {
	Direction    string      `json:"direction"`
	Strength     string      `json:"strength,omitempty"`
	Confidence   float64     `json:"confidence,omitempty"`
	Components   []Component `json:"components,omitempty"`
	SourceRef    string      `json:"source_ref"`
	Status       FieldStatus `json:"status"`
	Reliability  Reliability `json:"reliability"`
}

type Volatility struct {
	Regime      string      `json:"regime"`
	ATR         *float64    `json:"atr,omitempty"`
	ATRPct      *float64    `json:"atr_pct,omitempty"`
	RealizedPct *float64    `json:"realized_pct,omitempty"`
	SourceRef   string      `json:"source_ref"`
	Status      FieldStatus `json:"status"`
	Reliability Reliability `json:"reliability"`
}

type Level struct {
	Type        string      `json:"type"`
	Price       float64     `json:"price"`
	DistancePct float64     `json:"distance_pct,omitempty"`
	SourceRef   string      `json:"source_ref"`
	Status      FieldStatus `json:"status"`
	Reliability Reliability `json:"reliability"`
}

type Signal struct {
	Name        string      `json:"name"`
	Direction   string      `json:"direction"`
	Strength    string      `json:"strength,omitempty"`
	Value       *float64    `json:"value,omitempty"`
	EvidenceRefs []string   `json:"evidence_refs,omitempty"`
	SourceRef   string      `json:"source_ref"`
	Status      FieldStatus `json:"status"`
	Reliability Reliability `json:"reliability"`
}

type RiskContext struct {
	Bid             *float64    `json:"bid,omitempty"`
	Ask             *float64    `json:"ask,omitempty"`
	Spread          *float64    `json:"spread,omitempty"`
	SpreadPct       *float64    `json:"spread_pct,omitempty"`
	Lot             int         `json:"lot,omitempty"`
	MinPriceIncrement float64   `json:"min_price_increment,omitempty"`
	OrderbookDepth  int         `json:"orderbook_depth,omitempty"`
	LiquidityStatus string      `json:"liquidity_status,omitempty"`
	SessionStatus   string      `json:"session_status,omitempty"`
	SourceRef       string      `json:"source_ref"`
	Status          FieldStatus `json:"status"`
	Reliability     Reliability `json:"reliability"`
}

type Component struct {
	Name        string      `json:"name"`
	Value       float64     `json:"value,omitempty"`
	SourceRef   string      `json:"source_ref"`
	Status      FieldStatus `json:"status"`
	Reliability Reliability `json:"reliability"`
}

type DataQuality struct {
	Overall           FieldStatus `json:"overall"`
	Completeness      float64     `json:"completeness"`
	FreshnessSeconds  int64       `json:"freshness_seconds,omitempty"`
	TrustedFields     []string    `json:"trusted_fields,omitempty"`
	PartialFields     []string    `json:"partial_fields,omitempty"`
	UnavailableFields []string    `json:"unavailable_fields,omitempty"`
	UnreliableFields  []string    `json:"unreliable_fields,omitempty"`
	SourceCounts      map[SourceClass]int `json:"source_counts,omitempty"`
}

type Warning struct {
	Code      string   `json:"code"`
	Message   string   `json:"message"`
	FieldPath string   `json:"field_path,omitempty"`
	SourceRefs []string `json:"source_refs,omitempty"`
}

type SourceTrace struct {
	ID          string       `json:"id"`
	SourceClass SourceClass `json:"source_class"`
	Tool        string      `json:"tool,omitempty"`
	Service     string      `json:"service,omitempty"`
	Symbol      string      `json:"symbol,omitempty"`
	Timeframe   string      `json:"timeframe,omitempty"`
	AsOf        *time.Time  `json:"as_of,omitempty"`
	Status      FieldStatus `json:"status"`
	Reliability Reliability `json:"reliability"`
	Notes       string      `json:"notes,omitempty"`
}
```

---

## JSON Example

This is a schema example, not a live market snapshot.

```json
{
  "contract_version": "hts.market_summary.v1",
  "success": true,
  "generated_at": "2026-04-27T00:25:00Z",
  "analysis_symbol": {
    "symbol": "RUS:NG1!",
    "exchange": "RUS",
    "ticker": "NG1!",
    "type": "futures",
    "is_continuous": true,
    "source_ref": "tv.chart_state",
    "status": "ok",
    "reliability": "trusted"
  },
  "execution_symbol": {
    "symbol": "SPBFUT:NGM6",
    "ticker": "NGM6",
    "class_code": "SPBFUT",
    "figi": "TINKOFF_EXAMPLE_FIGI",
    "instrument_uid": "tinkoff-example-instrument-uid",
    "lot": 1,
    "min_price_increment": 0.001,
    "mapped_from": "RUS:NG1!",
    "source_ref": "tinkoff.instrument",
    "status": "ok",
    "reliability": "trusted"
  },
  "timeframe": {
    "value": "1D",
    "seconds": 86400,
    "source_ref": "tv.chart_state",
    "status": "ok",
    "reliability": "trusted"
  },
  "current_price": {
    "last": 3.125,
    "bid": 3.124,
    "ask": 3.126,
    "mid": 3.125,
    "currency": "USD",
    "as_of": "2026-04-27T00:24:58Z",
    "source_ref": "tinkoff.orderbook",
    "status": "ok",
    "reliability": "trusted"
  },
  "ohlcv_summary": {
    "bar_count": 120,
    "period_from": "2025-11-03T00:00:00Z",
    "period_to": "2026-04-24T00:00:00Z",
    "last_bar": {
      "time": "2026-04-24T00:00:00Z",
      "open": 3.08,
      "high": 3.16,
      "low": 3.02,
      "close": 3.12,
      "volume": 14200
    },
    "change": 0.04,
    "change_pct": 1.3,
    "average_volume": 11800,
    "volume_vs_avg": 1.2,
    "source_ref": "tv.ohlcv",
    "status": "ok",
    "reliability": "trusted"
  },
  "phase": {
    "value": "pullback_in_uptrend",
    "confidence": 0.68,
    "evidence_refs": ["hts.trend", "hts.volatility"],
    "source_ref": "hts.phase",
    "status": "ok",
    "reliability": "trusted"
  },
  "trend": {
    "direction": "up",
    "strength": "moderate",
    "confidence": 0.71,
    "components": [
      {
        "name": "ema_slope_20",
        "value": 0.018,
        "source_ref": "hts.ema20",
        "status": "ok",
        "reliability": "trusted"
      },
      {
        "name": "adx",
        "value": 24.6,
        "source_ref": "pine.hts_json",
        "status": "ok",
        "reliability": "trusted"
      }
    ],
    "source_ref": "hts.trend",
    "status": "ok",
    "reliability": "trusted"
  },
  "volatility": {
    "regime": "normal",
    "atr": 0.15,
    "atr_pct": 4.8,
    "realized_pct": 3.9,
    "source_ref": "hts.volatility",
    "status": "ok",
    "reliability": "trusted"
  },
  "levels": [
    {
      "type": "support",
      "price": 3.02,
      "distance_pct": -3.2,
      "source_ref": "hts.levels",
      "status": "ok",
      "reliability": "trusted"
    },
    {
      "type": "pine_resistance",
      "price": 3.27,
      "distance_pct": 4.6,
      "source_ref": "pine.hts_json",
      "status": "ok",
      "reliability": "trusted"
    }
  ],
  "signals": [
    {
      "name": "rsi_regime",
      "direction": "neutral_bullish",
      "strength": "weak",
      "value": 58.4,
      "evidence_refs": ["pine.hts_json"],
      "source_ref": "pine.hts_json",
      "status": "ok",
      "reliability": "trusted"
    },
    {
      "name": "data_window_indicator_snapshot",
      "direction": "unknown",
      "strength": "none",
      "source_ref": "tv.data_window_text",
      "status": "unreliable",
      "reliability": "unreliable"
    }
  ],
  "risk_context": {
    "bid": 3.124,
    "ask": 3.126,
    "spread": 0.002,
    "spread_pct": 0.064,
    "lot": 1,
    "min_price_increment": 0.001,
    "orderbook_depth": 20,
    "liquidity_status": "tradable",
    "session_status": "open",
    "source_ref": "tinkoff.orderbook",
    "status": "ok",
    "reliability": "trusted"
  },
  "data_quality": {
    "overall": "partial",
    "completeness": 0.86,
    "freshness_seconds": 2,
    "trusted_fields": [
      "analysis_symbol",
      "execution_symbol",
      "timeframe",
      "current_price",
      "ohlcv_summary",
      "risk_context"
    ],
    "partial_fields": ["phase", "trend", "volatility", "levels", "signals"],
    "unavailable_fields": [],
    "unreliable_fields": ["signals[1]"],
    "source_counts": {
      "tradingview_direct": 2,
      "hts_go_derived": 5,
      "pine_hts_json": 3,
      "tinkoff_marketdata": 2,
      "unreliable": 1
    }
  },
  "warnings": [
    {
      "code": "CONTINUOUS_ANALYSIS_SYMBOL",
      "message": "analysis_symbol is a TradingView continuous contract and must not be used for execution.",
      "field_path": "analysis_symbol",
      "source_refs": ["tv.chart_state"]
    },
    {
      "code": "UNRELIABLE_SIGNAL_SUPPRESSED",
      "message": "Data Window text snapshot is included only for diagnostics and must not be used as factual indicator value.",
      "field_path": "signals[1]",
      "source_refs": ["tv.data_window_text"]
    }
  ],
  "recommended_recheck_at": "2026-04-27T00:26:00Z",
  "source_trace": [
    {
      "id": "tv.chart_state",
      "source_class": "tradingview_direct",
      "tool": "chart_get_state",
      "symbol": "RUS:NG1!",
      "timeframe": "1D",
      "as_of": "2026-04-27T00:24:55Z",
      "status": "ok",
      "reliability": "trusted"
    },
    {
      "id": "tv.ohlcv",
      "source_class": "tradingview_direct",
      "tool": "data_get_ohlcv",
      "symbol": "RUS:NG1!",
      "timeframe": "1D",
      "as_of": "2026-04-27T00:24:55Z",
      "status": "ok",
      "reliability": "trusted"
    },
    {
      "id": "pine.hts_json",
      "source_class": "pine_hts_json",
      "tool": "data_get_pine_tables",
      "symbol": "RUS:NG1!",
      "timeframe": "1D",
      "as_of": "2026-04-27T00:24:56Z",
      "status": "ok",
      "reliability": "trusted",
      "notes": "schema=HTS_JSON/v1 script_hash=example"
    },
    {
      "id": "tinkoff.instrument",
      "source_class": "tinkoff_marketdata",
      "service": "MarketDataService/GetInstrumentBy",
      "symbol": "SPBFUT:NGM6",
      "as_of": "2026-04-27T00:24:58Z",
      "status": "ok",
      "reliability": "trusted"
    },
    {
      "id": "tinkoff.orderbook",
      "source_class": "tinkoff_marketdata",
      "service": "MarketDataService/GetOrderBook",
      "symbol": "SPBFUT:NGM6",
      "as_of": "2026-04-27T00:24:58Z",
      "status": "ok",
      "reliability": "trusted"
    },
    {
      "id": "tv.data_window_text",
      "source_class": "unreliable",
      "tool": "data_get_indicator",
      "symbol": "RUS:NG1!",
      "timeframe": "1D",
      "as_of": "2026-04-27T00:24:56Z",
      "status": "unreliable",
      "reliability": "unreliable",
      "notes": "localized text parsing and UI rounding are not safe for numerical decisions"
    },
    {
      "id": "hts.trend",
      "source_class": "hts_go_derived",
      "service": "HTS Go",
      "symbol": "RUS:NG1!",
      "timeframe": "1D",
      "as_of": "2026-04-27T00:24:59Z",
      "status": "ok",
      "reliability": "trusted",
      "notes": "derived from tv.ohlcv and pine.hts_json"
    },
    {
      "id": "hts.phase",
      "source_class": "hts_go_derived",
      "service": "HTS Go",
      "symbol": "RUS:NG1!",
      "timeframe": "1D",
      "as_of": "2026-04-27T00:24:59Z",
      "status": "ok",
      "reliability": "trusted"
    },
    {
      "id": "hts.volatility",
      "source_class": "hts_go_derived",
      "service": "HTS Go",
      "symbol": "RUS:NG1!",
      "timeframe": "1D",
      "as_of": "2026-04-27T00:24:59Z",
      "status": "ok",
      "reliability": "trusted"
    },
    {
      "id": "hts.levels",
      "source_class": "hts_go_derived",
      "service": "HTS Go",
      "symbol": "RUS:NG1!",
      "timeframe": "1D",
      "as_of": "2026-04-27T00:24:59Z",
      "status": "ok",
      "reliability": "trusted"
    },
    {
      "id": "hts.ema20",
      "source_class": "hts_go_derived",
      "service": "HTS Go",
      "symbol": "RUS:NG1!",
      "timeframe": "1D",
      "as_of": "2026-04-27T00:24:59Z",
      "status": "ok",
      "reliability": "trusted"
    }
  ]
}
```

---

## Fields That Must Not Be Sent To The LLM As Truth

The LLM may receive these fields only as diagnostics, with warnings, or not at
all:

- Any field where `status` is `unavailable`, `stale`, `unreliable`, or `error`.
- Any field where `reliability` is not `trusted`.
- Canvas pixels, canvas hit-test coordinates, Y coordinates, or screenshot
  derived values.
- Localized Data Window or DOM text parsed as numbers unless the parser and
  source are explicitly verified for the current locale and field.
- TradingView `bid`/`ask` values that are missing and represented as `0`.
- MOEX futures bid/ask/spread from TradingView when internal quote state does
  not expose these keys.
- `execution_symbol` when it was copied from a TradingView continuous contract
  instead of resolved to a real Tinkoff instrument.
- Empty strategy metrics, empty trade lists, or equity summaries returned with
  no loaded strategy.
- Pine output that is not strict `HTS_JSON`, has an unknown schema version, or
  lacks a verified script hash/name.
- `phase`, `trend`, `volatility`, `levels`, and `signals` without source refs
  and evidence refs.
- Any field older than `recommended_recheck_at`.

LLM prompt builders should filter or downgrade all paths listed in
`data_quality.unreliable_fields`, and should include `warnings` before asking
for interpretation.
