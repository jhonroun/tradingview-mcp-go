# LLM Market Context Contract

Status: draft, documentation only.

This contract defines compact market-analysis payloads for an LLM. The LLM
receives summaries, calculated features, quality flags, and warnings. It does
not receive raw candles unless a separate debug/research workflow explicitly
requests them.

Boundary: this document describes the external HTS MCP -> LLM interface. It
does not add trading execution, broker integration, risk sizing, or decision
automation to `tradingview-mcp-go`.

Related upstream contract:
[HTS_MARKET_SUMMARY_CONTRACT.md](HTS_MARKET_SUMMARY_CONTRACT.md).
For analysis-symbol to execution-instrument mapping, see
[INSTRUMENT_RESOLVER_CONTRACT.md](INSTRUMENT_RESOLVER_CONTRACT.md).

---

## Principles

- Default to compact features, not raw OHLCV arrays.
- Every field used for reasoning must expose `quality` and source references.
- LLM output must be structured and machine-checkable.
- LLM must not turn unavailable or unreliable data into factual statements.
- LLM may explain, compare, and request rechecks; it must not execute trades.
- Raw candles are allowed only in an explicit diagnostic payload with
  `raw_candles_included=true` and a warning.

---

## Context Types

### 1. Minimal Summary For One Instrument

Use this for a single chart read, market brief, signal check, or pre-trade
analysis.

Required payload fields:

- `instrument`: analysis and execution symbol refs.
- `timeframe`: timeframe and seconds.
- `snapshot`: current price, phase, trend, volatility, risk, and freshness.
- `features`: normalized calculated features, not raw bars.
- `levels`: compact support/resistance/zones.
- `signals`: compact signal states.
- `quality`: payload-level quality flags.
- `warnings`: field-specific warning flags.

Recommended features:

- `price_change_pct_1_bar`
- `price_change_pct_n_bars`
- `range_pct_n_bars`
- `volume_vs_avg`
- `atr_pct`
- `ema_fast_slope`
- `ema_slow_slope`
- `distance_to_nearest_support_pct`
- `distance_to_nearest_resistance_pct`
- `rsi`
- `adx`

### 2. Summary For A List Of Instruments

Use this for market scan and ranking. Each item is much smaller than the single
instrument summary.

Required payload fields:

- `universe`: name, filters, and count.
- `ranking`: score definition and sort order.
- `items`: compact per-instrument rows.
- `quality`: scan-level quality.
- `warnings`: scan-level and per-item warnings.

Each item should include:

- `analysis_symbol`
- `execution_symbol`
- `timeframe`
- `last`
- `phase`
- `trend`
- `volatility_regime`
- `risk_status`
- `score`
- `top_signals`
- `blocking_warnings`

### 3. Diff Summary For Re-Evaluation

Use this when the LLM already analyzed a previous context and HTS needs a cheap
recheck.

Required payload fields:

- `baseline_context_id`
- `current_context_id`
- `elapsed_seconds`
- `changed_fields`
- `material_changes`
- `unchanged_fields`
- `quality_delta`
- `recheck_reason`

Material changes should focus on:

- price move beyond threshold;
- trend/phase change;
- volatility regime change;
- nearest level crossed or approached;
- spread/liquidity degraded;
- signal flip;
- source quality degraded.

### 4. Review Summary For Trade Analysis

Use this after a planned or executed trade needs explanation. Do not include raw
tick/candle history by default.

Required payload fields:

- `trade`: side, planned entry/stop/target, actual entry/exit when known.
- `pre_trade_context`: compact context at decision time.
- `post_trade_context`: compact context at review time.
- `outcome`: pnl, r-multiple, max favorable/adverse excursion if available.
- `attribution`: structured factors that helped or hurt.
- `rule_checks`: strategy/risk rule pass/fail list.
- `quality`: review quality.
- `warnings`: warning flags.

### 5. Warning Flags

Warning flags are first-class data. The prompt builder must include them before
asking the model for interpretation.

Core warning codes:

- `RAW_CANDLES_INCLUDED`: raw candles were included; model must not overfit.
- `STALE_CONTEXT`: context is older than the allowed freshness window.
- `LOW_COMPLETENESS`: required fields are missing or partial.
- `UNTRUSTED_SOURCE`: at least one important field is not trusted.
- `UNRELIABLE_FIELD_PRESENT`: unreliable diagnostic field is present.
- `CANVAS_COORDINATES_PRESENT`: canvas/pixel coordinate data is present.
- `DOM_TEXT_NUMERIC_PRESENT`: numeric data came from DOM/Data Window text.
- `TV_BID_ASK_ZERO_SENTINEL`: TradingView bid/ask is `0` because unavailable.
- `BID_ASK_UNAVAILABLE`: executable bid/ask/spread is unavailable.
- `CONTINUOUS_CONTRACT_ANALYSIS_ONLY`: analysis symbol is continuous futures.
- `EXECUTION_SYMBOL_UNRESOLVED`: no real tradable instrument was resolved.
- `TIMEFRAME_MISMATCH`: sources use different timeframes.
- `SOURCE_STALE`: one or more sources are stale.
- `PINE_JSON_UNVERIFIED`: Pine machine-readable output is not verified.
- `STRATEGY_METRICS_EMPTY`: strategy metrics/trades/equity are empty.
- `DIFF_BASELINE_MISMATCH`: diff baseline does not match current instrument.
- `MODEL_OUTPUT_SCHEMA_REQUIRED`: LLM must return strict JSON only.

### 6. LLM Response Format

LLM response must be JSON and must include:

- `schema_version`
- `context_id`
- `task_type`
- `verdict`
- `confidence`
- `summary`
- `key_drivers`
- `risks`
- `invalidations`
- `rechecks`
- `referenced_warnings`
- `uses_unreliable_fields`
- `prohibited_actions_detected`

Allowed verdict values:

- `insufficient_data`
- `no_action`
- `watch`
- `bullish_bias`
- `bearish_bias`
- `range_bound`
- `risk_too_high`
- `review_only`

The LLM response is analysis, not an execution command.

---

## JSON Schema

This schema is intentionally compact. It defines the LLM-facing envelope, four
payload types, shared quality/warning objects, and the required LLM response.

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://tradingview-mcp-go.local/schemas/llm_market_context.v1.json",
  "title": "LLM Market Context Contract",
  "type": "object",
  "required": [
    "schema_version",
    "context_id",
    "context_type",
    "generated_at",
    "raw_candles_included",
    "quality",
    "warnings",
    "payload"
  ],
  "properties": {
    "schema_version": {
      "const": "hts.llm_context.v1"
    },
    "context_id": {
      "type": "string",
      "minLength": 1
    },
    "context_type": {
      "enum": [
        "instrument_summary",
        "market_scan_summary",
        "diff_summary",
        "trade_review_summary"
      ]
    },
    "generated_at": {
      "type": "string",
      "format": "date-time"
    },
    "expires_at": {
      "type": "string",
      "format": "date-time"
    },
    "raw_candles_included": {
      "type": "boolean",
      "default": false
    },
    "token_budget": {
      "$ref": "#/$defs/token_budget"
    },
    "quality": {
      "$ref": "#/$defs/quality"
    },
    "warnings": {
      "type": "array",
      "items": {
        "$ref": "#/$defs/warning"
      }
    },
    "payload": {
      "oneOf": [
        {
          "$ref": "#/$defs/instrument_summary"
        },
        {
          "$ref": "#/$defs/market_scan_summary"
        },
        {
          "$ref": "#/$defs/diff_summary"
        },
        {
          "$ref": "#/$defs/trade_review_summary"
        }
      ]
    }
  },
  "$defs": {
    "token_budget": {
      "type": "object",
      "required": ["target_tokens", "max_tokens"],
      "properties": {
        "target_tokens": { "type": "integer", "minimum": 1 },
        "max_tokens": { "type": "integer", "minimum": 1 }
      }
    },
    "quality": {
      "type": "object",
      "required": ["overall", "completeness", "freshness_seconds", "trusted_for_llm"],
      "properties": {
        "overall": {
          "enum": ["ok", "partial", "stale", "unreliable", "error"]
        },
        "completeness": {
          "type": "number",
          "minimum": 0,
          "maximum": 1
        },
        "freshness_seconds": {
          "type": "integer",
          "minimum": 0
        },
        "trusted_for_llm": {
          "type": "boolean"
        },
        "trusted_fields": {
          "type": "array",
          "items": { "type": "string" }
        },
        "partial_fields": {
          "type": "array",
          "items": { "type": "string" }
        },
        "unreliable_fields": {
          "type": "array",
          "items": { "type": "string" }
        }
      }
    },
    "warning": {
      "type": "object",
      "required": ["code", "severity", "message"],
      "properties": {
        "code": { "type": "string" },
        "severity": {
          "enum": ["info", "warning", "critical"]
        },
        "message": { "type": "string" },
        "field_path": { "type": "string" },
        "source_refs": {
          "type": "array",
          "items": { "type": "string" }
        }
      }
    },
    "source_ref": {
      "type": "object",
      "required": ["id", "source_class", "status", "reliability"],
      "properties": {
        "id": { "type": "string" },
        "source_class": {
          "enum": [
            "tradingview_direct",
            "hts_go_derived",
            "pine_hts_json",
            "tinkoff_marketdata",
            "unreliable"
          ]
        },
        "status": {
          "enum": ["ok", "partial", "unavailable", "stale", "unreliable", "error"]
        },
        "reliability": {
          "enum": [
            "trusted",
            "usable_with_warning",
            "unavailable",
            "stale",
            "unreliable"
          ]
        }
      }
    },
    "symbol_ref": {
      "type": "object",
      "required": ["symbol", "source_ref"],
      "properties": {
        "symbol": { "type": "string" },
        "exchange": { "type": "string" },
        "ticker": { "type": "string" },
        "is_continuous": { "type": "boolean" },
        "source_ref": { "type": "string" }
      }
    },
    "execution_symbol": {
      "type": "object",
      "required": ["status", "source_ref"],
      "properties": {
        "status": {
          "enum": ["ok", "unresolved", "unavailable"]
        },
        "symbol": { "type": "string" },
        "figi": { "type": "string" },
        "instrument_uid": { "type": "string" },
        "lot": { "type": "integer" },
        "min_price_increment": { "type": "number" },
        "source_ref": { "type": "string" }
      }
    },
    "feature": {
      "type": "object",
      "required": ["name", "value", "quality", "source_ref"],
      "properties": {
        "name": { "type": "string" },
        "value": {
          "type": ["number", "string", "boolean", "null"]
        },
        "unit": { "type": "string" },
        "quality": {
          "enum": ["trusted", "usable_with_warning", "unavailable", "stale", "unreliable"]
        },
        "source_ref": { "type": "string" }
      }
    },
    "level": {
      "type": "object",
      "required": ["type", "price", "quality", "source_ref"],
      "properties": {
        "type": {
          "enum": ["support", "resistance", "zone_high", "zone_low", "vwap", "manual", "pine"]
        },
        "price": { "type": "number" },
        "distance_pct": { "type": "number" },
        "quality": {
          "enum": ["trusted", "usable_with_warning", "unavailable", "stale", "unreliable"]
        },
        "source_ref": { "type": "string" }
      }
    },
    "signal": {
      "type": "object",
      "required": ["name", "state", "quality", "source_ref"],
      "properties": {
        "name": { "type": "string" },
        "state": {
          "enum": ["bullish", "bearish", "neutral", "mixed", "blocked", "unknown"]
        },
        "strength": {
          "enum": ["none", "weak", "moderate", "strong"]
        },
        "value": {
          "type": ["number", "string", "boolean", "null"]
        },
        "quality": {
          "enum": ["trusted", "usable_with_warning", "unavailable", "stale", "unreliable"]
        },
        "source_ref": { "type": "string" }
      }
    },
    "risk_snapshot": {
      "type": "object",
      "required": ["status", "quality", "source_ref"],
      "properties": {
        "status": {
          "enum": ["tradable", "watch_only", "blocked", "unknown"]
        },
        "bid": { "type": ["number", "null"] },
        "ask": { "type": ["number", "null"] },
        "spread_pct": { "type": ["number", "null"] },
        "session_status": { "type": "string" },
        "quality": {
          "enum": ["trusted", "usable_with_warning", "unavailable", "stale", "unreliable"]
        },
        "source_ref": { "type": "string" }
      }
    },
    "instrument_summary": {
      "type": "object",
      "required": [
        "instrument",
        "timeframe",
        "snapshot",
        "features",
        "levels",
        "signals",
        "source_refs"
      ],
      "properties": {
        "instrument": {
          "type": "object",
          "required": ["analysis_symbol", "execution_symbol"],
          "properties": {
            "analysis_symbol": { "$ref": "#/$defs/symbol_ref" },
            "execution_symbol": { "$ref": "#/$defs/execution_symbol" }
          }
        },
        "timeframe": {
          "type": "object",
          "required": ["value"],
          "properties": {
            "value": { "type": "string" },
            "seconds": { "type": "integer" }
          }
        },
        "snapshot": {
          "type": "object",
          "required": ["last", "phase", "trend", "volatility", "risk"],
          "properties": {
            "last": { "type": "number" },
            "phase": { "type": "string" },
            "trend": { "type": "string" },
            "volatility": { "type": "string" },
            "risk": { "$ref": "#/$defs/risk_snapshot" }
          }
        },
        "features": {
          "type": "array",
          "items": { "$ref": "#/$defs/feature" }
        },
        "levels": {
          "type": "array",
          "items": { "$ref": "#/$defs/level" }
        },
        "signals": {
          "type": "array",
          "items": { "$ref": "#/$defs/signal" }
        },
        "source_refs": {
          "type": "array",
          "items": { "$ref": "#/$defs/source_ref" }
        }
      }
    },
    "market_scan_summary": {
      "type": "object",
      "required": ["universe", "ranking", "items"],
      "properties": {
        "universe": {
          "type": "object",
          "required": ["name", "count"],
          "properties": {
            "name": { "type": "string" },
            "count": { "type": "integer" },
            "filters": {
              "type": "array",
              "items": { "type": "string" }
            }
          }
        },
        "ranking": {
          "type": "object",
          "required": ["score_name", "sort"],
          "properties": {
            "score_name": { "type": "string" },
            "sort": { "enum": ["asc", "desc"] }
          }
        },
        "items": {
          "type": "array",
          "items": {
            "type": "object",
            "required": [
              "analysis_symbol",
              "timeframe",
              "last",
              "phase",
              "trend",
              "volatility_regime",
              "score",
              "top_signals",
              "blocking_warnings"
            ],
            "properties": {
              "analysis_symbol": { "type": "string" },
              "execution_symbol_status": { "type": "string" },
              "timeframe": { "type": "string" },
              "last": { "type": "number" },
              "phase": { "type": "string" },
              "trend": { "type": "string" },
              "volatility_regime": { "type": "string" },
              "risk_status": { "type": "string" },
              "score": { "type": "number" },
              "top_signals": {
                "type": "array",
                "items": { "type": "string" }
              },
              "blocking_warnings": {
                "type": "array",
                "items": { "type": "string" }
              }
            }
          }
        }
      }
    },
    "diff_summary": {
      "type": "object",
      "required": [
        "baseline_context_id",
        "current_context_id",
        "elapsed_seconds",
        "changed_fields",
        "material_changes",
        "recheck_reason"
      ],
      "properties": {
        "baseline_context_id": { "type": "string" },
        "current_context_id": { "type": "string" },
        "elapsed_seconds": { "type": "integer" },
        "changed_fields": {
          "type": "array",
          "items": { "type": "string" }
        },
        "material_changes": {
          "type": "array",
          "items": { "$ref": "#/$defs/feature" }
        },
        "unchanged_fields": {
          "type": "array",
          "items": { "type": "string" }
        },
        "quality_delta": {
          "type": "string"
        },
        "recheck_reason": {
          "type": "string"
        }
      }
    },
    "trade_review_summary": {
      "type": "object",
      "required": ["trade", "pre_trade_context", "post_trade_context", "outcome", "rule_checks"],
      "properties": {
        "trade": {
          "type": "object",
          "required": ["side", "planned_entry", "planned_stop"],
          "properties": {
            "side": { "enum": ["long", "short"] },
            "planned_entry": { "type": "number" },
            "planned_stop": { "type": "number" },
            "planned_target": { "type": ["number", "null"] },
            "actual_entry": { "type": ["number", "null"] },
            "actual_exit": { "type": ["number", "null"] }
          }
        },
        "pre_trade_context": { "$ref": "#/$defs/instrument_summary" },
        "post_trade_context": { "$ref": "#/$defs/instrument_summary" },
        "outcome": {
          "type": "object",
          "properties": {
            "pnl": { "type": ["number", "null"] },
            "r_multiple": { "type": ["number", "null"] },
            "max_favorable_excursion_r": { "type": ["number", "null"] },
            "max_adverse_excursion_r": { "type": ["number", "null"] }
          }
        },
        "attribution": {
          "type": "array",
          "items": { "$ref": "#/$defs/feature" }
        },
        "rule_checks": {
          "type": "array",
          "items": {
            "type": "object",
            "required": ["rule", "status"],
            "properties": {
              "rule": { "type": "string" },
              "status": { "enum": ["pass", "fail", "unknown"] },
              "evidence": { "type": "string" }
            }
          }
        }
      }
    },
    "llm_response": {
      "type": "object",
      "required": [
        "schema_version",
        "context_id",
        "task_type",
        "verdict",
        "confidence",
        "summary",
        "key_drivers",
        "risks",
        "invalidations",
        "rechecks",
        "referenced_warnings",
        "uses_unreliable_fields",
        "prohibited_actions_detected"
      ],
      "properties": {
        "schema_version": { "const": "hts.llm_response.v1" },
        "context_id": { "type": "string" },
        "task_type": {
          "enum": [
            "instrument_analysis",
            "market_scan",
            "diff_recheck",
            "trade_review"
          ]
        },
        "verdict": {
          "enum": [
            "insufficient_data",
            "no_action",
            "watch",
            "bullish_bias",
            "bearish_bias",
            "range_bound",
            "risk_too_high",
            "review_only"
          ]
        },
        "confidence": { "type": "number", "minimum": 0, "maximum": 1 },
        "summary": { "type": "string" },
        "key_drivers": {
          "type": "array",
          "items": { "type": "string" }
        },
        "risks": {
          "type": "array",
          "items": { "type": "string" }
        },
        "invalidations": {
          "type": "array",
          "items": { "type": "string" }
        },
        "rechecks": {
          "type": "array",
          "items": { "type": "string" }
        },
        "referenced_warnings": {
          "type": "array",
          "items": { "type": "string" }
        },
        "uses_unreliable_fields": { "type": "boolean" },
        "prohibited_actions_detected": {
          "type": "array",
          "items": { "type": "string" }
        }
      }
    }
  }
}
```

---

## Proposed Go Structs

These structs are proposed for the external HTS MCP layer. They are not
implemented in `tradingview-mcp-go`.

```go
package llmcontext

import "time"

type ContextType string

const (
	ContextInstrumentSummary ContextType = "instrument_summary"
	ContextMarketScanSummary ContextType = "market_scan_summary"
	ContextDiffSummary       ContextType = "diff_summary"
	ContextTradeReview       ContextType = "trade_review_summary"
)

type Reliability string

const (
	ReliabilityTrusted           Reliability = "trusted"
	ReliabilityUsableWithWarning Reliability = "usable_with_warning"
	ReliabilityUnavailable       Reliability = "unavailable"
	ReliabilityStale             Reliability = "stale"
	ReliabilityUnreliable        Reliability = "unreliable"
)

type LLMContextEnvelope struct {
	SchemaVersion     string       `json:"schema_version"`
	ContextID         string       `json:"context_id"`
	ContextType       ContextType  `json:"context_type"`
	GeneratedAt       time.Time    `json:"generated_at"`
	ExpiresAt         *time.Time   `json:"expires_at,omitempty"`
	RawCandlesIncluded bool        `json:"raw_candles_included"`
	TokenBudget       TokenBudget  `json:"token_budget"`
	Quality           Quality      `json:"quality"`
	Warnings          []Warning    `json:"warnings"`
	Payload           any          `json:"payload"`
}

type TokenBudget struct {
	TargetTokens int `json:"target_tokens"`
	MaxTokens    int `json:"max_tokens"`
}

type Quality struct {
	Overall          string   `json:"overall"`
	Completeness     float64  `json:"completeness"`
	FreshnessSeconds int64    `json:"freshness_seconds"`
	TrustedForLLM    bool     `json:"trusted_for_llm"`
	TrustedFields    []string `json:"trusted_fields,omitempty"`
	PartialFields    []string `json:"partial_fields,omitempty"`
	UnreliableFields []string `json:"unreliable_fields,omitempty"`
}

type Warning struct {
	Code       string   `json:"code"`
	Severity   string   `json:"severity"`
	Message    string   `json:"message"`
	FieldPath  string   `json:"field_path,omitempty"`
	SourceRefs []string `json:"source_refs,omitempty"`
}

type SourceRef struct {
	ID          string      `json:"id"`
	SourceClass string     `json:"source_class"`
	Status      string     `json:"status"`
	Reliability Reliability `json:"reliability"`
}

type SymbolRef struct {
	Symbol       string `json:"symbol"`
	Exchange     string `json:"exchange,omitempty"`
	Ticker       string `json:"ticker,omitempty"`
	IsContinuous bool   `json:"is_continuous,omitempty"`
	SourceRef    string `json:"source_ref"`
}

type ExecutionSymbolRef struct {
	Status            string  `json:"status"`
	Symbol            string  `json:"symbol,omitempty"`
	FIGI              string  `json:"figi,omitempty"`
	InstrumentUID     string  `json:"instrument_uid,omitempty"`
	Lot               int     `json:"lot,omitempty"`
	MinPriceIncrement float64 `json:"min_price_increment,omitempty"`
	SourceRef         string  `json:"source_ref"`
}

type TimeframeRef struct {
	Value   string `json:"value"`
	Seconds int64  `json:"seconds,omitempty"`
}

type Feature struct {
	Name       string      `json:"name"`
	Value      any         `json:"value"`
	Unit       string      `json:"unit,omitempty"`
	Quality    Reliability `json:"quality"`
	SourceRef  string      `json:"source_ref"`
}

type Level struct {
	Type        string      `json:"type"`
	Price       float64     `json:"price"`
	DistancePct float64     `json:"distance_pct,omitempty"`
	Quality     Reliability `json:"quality"`
	SourceRef   string      `json:"source_ref"`
}

type Signal struct {
	Name      string      `json:"name"`
	State     string      `json:"state"`
	Strength  string      `json:"strength,omitempty"`
	Value     any         `json:"value,omitempty"`
	Quality   Reliability `json:"quality"`
	SourceRef string      `json:"source_ref"`
}

type RiskSnapshot struct {
	Status        string      `json:"status"`
	Bid           *float64    `json:"bid,omitempty"`
	Ask           *float64    `json:"ask,omitempty"`
	SpreadPct     *float64    `json:"spread_pct,omitempty"`
	SessionStatus string      `json:"session_status,omitempty"`
	Quality       Reliability `json:"quality"`
	SourceRef     string      `json:"source_ref"`
}

type InstrumentSummaryPayload struct {
	Instrument struct {
		AnalysisSymbol  SymbolRef          `json:"analysis_symbol"`
		ExecutionSymbol ExecutionSymbolRef `json:"execution_symbol"`
	} `json:"instrument"`
	Timeframe TimeframeRef `json:"timeframe"`
	Snapshot  struct {
		Last       float64      `json:"last"`
		Phase      string       `json:"phase"`
		Trend      string       `json:"trend"`
		Volatility string       `json:"volatility"`
		Risk       RiskSnapshot `json:"risk"`
	} `json:"snapshot"`
	Features   []Feature   `json:"features"`
	Levels     []Level     `json:"levels"`
	Signals    []Signal    `json:"signals"`
	SourceRefs []SourceRef `json:"source_refs"`
}

type MarketScanPayload struct {
	Universe struct {
		Name    string   `json:"name"`
		Count   int      `json:"count"`
		Filters []string `json:"filters,omitempty"`
	} `json:"universe"`
	Ranking struct {
		ScoreName string `json:"score_name"`
		Sort      string `json:"sort"`
	} `json:"ranking"`
	Items []ScanItem `json:"items"`
}

type ScanItem struct {
	AnalysisSymbol        string   `json:"analysis_symbol"`
	ExecutionSymbolStatus string   `json:"execution_symbol_status,omitempty"`
	Timeframe             string   `json:"timeframe"`
	Last                  float64  `json:"last"`
	Phase                 string   `json:"phase"`
	Trend                 string   `json:"trend"`
	VolatilityRegime      string   `json:"volatility_regime"`
	RiskStatus            string   `json:"risk_status,omitempty"`
	Score                 float64  `json:"score"`
	TopSignals            []string `json:"top_signals"`
	BlockingWarnings      []string `json:"blocking_warnings"`
}

type DiffSummaryPayload struct {
	BaselineContextID string    `json:"baseline_context_id"`
	CurrentContextID  string    `json:"current_context_id"`
	ElapsedSeconds    int64     `json:"elapsed_seconds"`
	ChangedFields     []string  `json:"changed_fields"`
	MaterialChanges   []Feature `json:"material_changes"`
	UnchangedFields   []string  `json:"unchanged_fields,omitempty"`
	QualityDelta      string    `json:"quality_delta,omitempty"`
	RecheckReason     string    `json:"recheck_reason"`
}

type TradeReviewPayload struct {
	Trade struct {
		Side          string   `json:"side"`
		PlannedEntry  float64  `json:"planned_entry"`
		PlannedStop   float64  `json:"planned_stop"`
		PlannedTarget *float64 `json:"planned_target,omitempty"`
		ActualEntry   *float64 `json:"actual_entry,omitempty"`
		ActualExit    *float64 `json:"actual_exit,omitempty"`
	} `json:"trade"`
	PreTradeContext  InstrumentSummaryPayload `json:"pre_trade_context"`
	PostTradeContext InstrumentSummaryPayload `json:"post_trade_context"`
	Outcome          TradeOutcome             `json:"outcome"`
	Attribution      []Feature                `json:"attribution,omitempty"`
	RuleChecks       []RuleCheck              `json:"rule_checks"`
}

type TradeOutcome struct {
	PNL                    *float64 `json:"pnl,omitempty"`
	RMultiple              *float64 `json:"r_multiple,omitempty"`
	MaxFavorableExcursionR *float64 `json:"max_favorable_excursion_r,omitempty"`
	MaxAdverseExcursionR   *float64 `json:"max_adverse_excursion_r,omitempty"`
}

type RuleCheck struct {
	Rule     string `json:"rule"`
	Status   string `json:"status"`
	Evidence string `json:"evidence,omitempty"`
}

type LLMResponse struct {
	SchemaVersion             string   `json:"schema_version"`
	ContextID                 string   `json:"context_id"`
	TaskType                  string   `json:"task_type"`
	Verdict                   string   `json:"verdict"`
	Confidence                float64  `json:"confidence"`
	Summary                   string   `json:"summary"`
	KeyDrivers                []string `json:"key_drivers"`
	Risks                     []string `json:"risks"`
	Invalidations             []string `json:"invalidations"`
	Rechecks                  []string `json:"rechecks"`
	ReferencedWarnings        []string `json:"referenced_warnings"`
	UsesUnreliableFields      bool     `json:"uses_unreliable_fields"`
	ProhibitedActionsDetected []string `json:"prohibited_actions_detected"`
}
```

---

## Prompt Templates

Use the same JSON contract for all providers. Keep provider-specific prompts
short; the schema and context carry the detail.

### DeepSeek

```text
You are a market analysis model. Use only the provided hts.llm_context.v1 JSON.
Do not infer from missing raw candles. Treat warning flags as constraints.
Think through the evidence internally, then return only hts.llm_response.v1 JSON.

Task: {{task_type}}
Context JSON:
{{context_json}}
```

### Qwen

```text
Analyze the provided compact market context. Do not request or reconstruct raw
candles. If quality.trusted_for_llm is false, return verdict "insufficient_data".
Return strict JSON matching hts.llm_response.v1, with no markdown.

Task: {{task_type}}
Context JSON:
{{context_json}}
```

### Kimi

```text
You will receive a compact HTS market context. The context intentionally omits
raw candles. Use features, levels, signals, quality, and warnings only. Do not
expand assumptions beyond the data. Return only valid JSON in hts.llm_response.v1.

Task: {{task_type}}
Context JSON:
{{context_json}}
```

### Opus

```text
Act as a cautious market reviewer. First apply data-quality gates, then analyze
only trusted or usable_with_warning fields. Mention warning codes by ID. Do not
produce orders, position sizes, or execution instructions. Return only strict
hts.llm_response.v1 JSON.

Task: {{task_type}}
Context JSON:
{{context_json}}
```

### ChatGPT

```text
Use the following compact HTS market context. Raw candles are not available by
design. Base the answer only on fields with trusted quality unless the output
explicitly marks a warning. Return JSON only, matching hts.llm_response.v1.

Task: {{task_type}}
Context JSON:
{{context_json}}
```

---

## Forbidden Practices

- Passing raw OHLCV arrays to the LLM by default.
- Asking the LLM to calculate indicators from raw candles when HTS can calculate
  them deterministically in Go.
- Passing canvas coordinates, pixel values, or screenshot-derived values as
  market facts.
- Passing localized DOM/Data Window text as numeric truth without parser and
  source verification.
- Treating TradingView `bid=0` or `ask=0` as a real market quote.
- Using a TradingView continuous futures contract as an executable symbol.
- Hiding missing source data behind `success=true` without quality flags.
- Asking the LLM to infer missing bid/ask, spread, volume, trades, or strategy
  metrics.
- Passing Pine output as trusted unless it is strict `HTS_JSON` with verified
  schema/version/script identity.
- Letting the LLM change strategy parameters, generate execution orders, or
  choose position size directly.
- Accepting free-form prose from the LLM when a downstream system expects JSON.
- Ignoring `recommended_recheck_at`, `expires_at`, `STALE_CONTEXT`, or
  `SOURCE_STALE`.
- Comparing diff summaries from different symbols, timeframes, or incompatible
  source sets.
- Removing warning flags to reduce token count.
