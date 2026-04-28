# Instrument Resolver Contract

Status: draft, documentation only.

This document defines the interface between a TradingView analysis symbol and a
Tinkoff execution instrument for an external HTS MCP layer.

Boundary: `tradingview-mcp-go` may provide TradingView chart state, continuous
contract context, OHLCV, screenshots, Pine output, and other analysis inputs.
It must not implement Tinkoff integration, broker execution, risk sizing, or
order management. The resolver described here belongs to the external HTS MCP.

Related contracts:

- [HTS_MARKET_SUMMARY_CONTRACT.md](HTS_MARKET_SUMMARY_CONTRACT.md)
- [LLM_MARKET_CONTEXT_CONTRACT.md](LLM_MARKET_CONTEXT_CONTRACT.md)

---

## Core Rule

TradingView symbols and Tinkoff instruments are different identities.

```text
TradingView continuous futures -> analysis only
Tinkoff concrete futures       -> execution, orderbook, expiration, margin, status
```

Never copy `RUS:NG1!`, `MOEX:NG1!`, `NYMEX:NG1!`, or any other continuous
TradingView symbol into an executable instrument field. A resolver must map the
analysis symbol to a concrete Tinkoff `instrument_uid` / FIGI before execution
data can be trusted.

---

## Responsibilities

The resolver is responsible for:

- normalizing a TradingView analysis symbol;
- detecting continuous futures and roll number;
- finding candidate concrete Tinkoff futures;
- selecting the execution instrument for the requested `as_of` time;
- returning bid/ask/spread from Tinkoff orderbook, not TradingView quote
  sentinels;
- returning expiration, last trade date, margin, lot, min price increment, and
  trading status from Tinkoff;
- refusing to resolve when candidates are missing, ambiguous, stale, or not
  tradable;
- recording resolver provenance so HTS and LLM layers can explain the mapping.

The resolver is not responsible for:

- placing orders;
- choosing position size;
- deciding whether a trade should be opened;
- backtesting;
- changing TradingView chart symbols.

---

## Resolution Flow

1. Parse `analysis_symbol`.
   - Example: `RUS:NG1!`
   - Exchange: `RUS`
   - Ticker: `NG1!`
   - Root: `NG`
   - Roll number: `1`
   - Continuous: `true`

2. Look for an active manual mapping.
   - Manual mapping wins only if it is not expired and the target Tinkoff
     instrument is still tradable.

3. Load or refresh Tinkoff futures catalog.
   - Use concrete instruments only.
   - Keep `instrument_uid`, FIGI, ticker, class code, expiration, first/last
     trade dates, lot, min price increment, currency, exchange, and status.

4. Find candidates through root mapping hints.
   - Example root mapping: TradingView `RUS:NG` -> Tinkoff `SPBFUT` root hint
     `NG`.
   - Tickers do not need to match exactly.

5. Filter candidates.
   - Expiration and last trade date must be after `as_of`.
   - Trading status must be compatible with execution.
   - Instrument must have usable market data if execution context is requested.

6. Select roll target.
   - `NG1!` normally maps to nearest tradable future by expiration.
   - `NG2!` normally maps to the second nearest tradable future.
   - Manual override may choose a different target with an explicit reason.

7. Validate execution data.
   - Get Tinkoff orderbook for bid/ask/spread.
   - Get Tinkoff futures margin/status.
   - If bid/ask is unavailable, return `watch_only` or `unresolved`, not fake
     executable data.

8. Return a structured `ResolutionResult`.

---

## Proposed Go Interfaces

These interfaces are proposed for the external HTS MCP layer. They are not
implemented in `tradingview-mcp-go`.

```go
package resolver

import (
	"context"
	"time"
)

type InstrumentResolver interface {
	Resolve(ctx context.Context, req ResolveRequest) (ResolutionResult, error)
	Candidates(ctx context.Context, req CandidateRequest) (CandidateResult, error)
	RefreshCatalog(ctx context.Context, req RefreshCatalogRequest) (RefreshCatalogResult, error)
	GetMapping(ctx context.Context, key MappingKey) (ManualMapping, error)
	UpsertMapping(ctx context.Context, mapping ManualMapping) (ManualMapping, error)
	ValidateMapping(ctx context.Context, mapping ManualMapping, asOf time.Time) (ValidationResult, error)
}

type TinkoffMarketData interface {
	FuturesCatalog(ctx context.Context) ([]TinkoffFuture, error)
	OrderBook(ctx context.Context, instrumentUID string, depth int) (OrderBookSnapshot, error)
	FuturesMargin(ctx context.Context, instrumentUID string) (FuturesMargin, error)
	TradingStatus(ctx context.Context, instrumentUID string) (InstrumentStatus, error)
}
```

---

## Proposed Go Structs

```go
package resolver

import "time"

type ResolveRequest struct {
	AnalysisSymbol     AnalysisSymbolRef `json:"analysis_symbol"`
	Timeframe          string            `json:"timeframe,omitempty"`
	AsOf               time.Time         `json:"as_of"`
	PreferredClassCode string            `json:"preferred_class_code,omitempty"`
	PreferredExchange  string            `json:"preferred_exchange,omitempty"`
	DesiredRoll        int               `json:"desired_roll,omitempty"`
	OrderBookDepth     int               `json:"orderbook_depth,omitempty"`
	AllowManualOnly    bool              `json:"allow_manual_only,omitempty"`
	RequireTradable    bool              `json:"require_tradable"`
}

type AnalysisSymbolRef struct {
	Provider      string `json:"provider"` // tradingview
	Raw           string `json:"raw"`      // RUS:NG1!
	Exchange      string `json:"exchange,omitempty"`
	Ticker        string `json:"ticker,omitempty"`
	Root          string `json:"root,omitempty"`
	RollNumber    int    `json:"roll_number,omitempty"`
	IsContinuous  bool   `json:"is_continuous"`
	SourceRef     string `json:"source_ref,omitempty"`
}

type ResolutionStatus string

const (
	ResolutionResolved       ResolutionStatus = "resolved"
	ResolutionManualRequired ResolutionStatus = "manual_required"
	ResolutionUnresolved     ResolutionStatus = "unresolved"
	ResolutionAmbiguous      ResolutionStatus = "ambiguous"
	ResolutionStale          ResolutionStatus = "stale"
	ResolutionNotTradable    ResolutionStatus = "not_tradable"
)

type ResolutionResult struct {
	Success          bool                `json:"success"`
	Status           ResolutionStatus    `json:"status"`
	AnalysisSymbol   AnalysisSymbolRef   `json:"analysis_symbol"`
	ExecutionSymbol  *ExecutionSymbolRef `json:"execution_symbol,omitempty"`
	SelectedBy       string              `json:"selected_by,omitempty"` // manual_mapping, auto_roll, cached
	Confidence       float64             `json:"confidence"`
	Reason           string              `json:"reason,omitempty"`
	Candidates        []ExecutionSymbolRef `json:"candidates,omitempty"`
	OrderBook         *OrderBookSnapshot  `json:"orderbook,omitempty"`
	Margin            *FuturesMargin      `json:"margin,omitempty"`
	InstrumentStatus  *InstrumentStatus  `json:"instrument_status,omitempty"`
	Warnings          []ResolverWarning  `json:"warnings,omitempty"`
	RecommendedAction string              `json:"recommended_action,omitempty"`
	ResolvedAt        time.Time           `json:"resolved_at"`
	ExpiresAt         time.Time           `json:"expires_at"`
	SourceTrace       []ResolverSource    `json:"source_trace,omitempty"`
}

type ExecutionSymbolRef struct {
	Provider          string     `json:"provider"` // tinkoff
	Symbol            string     `json:"symbol,omitempty"`
	Ticker            string     `json:"ticker,omitempty"`
	ClassCode         string     `json:"class_code,omitempty"`
	Exchange          string     `json:"exchange,omitempty"`
	FIGI              string     `json:"figi,omitempty"`
	InstrumentUID     string     `json:"instrument_uid"`
	AssetUID          string     `json:"asset_uid,omitempty"`
	PositionUID       string     `json:"position_uid,omitempty"`
	Lot               int        `json:"lot,omitempty"`
	MinPriceIncrement Decimal    `json:"min_price_increment,omitempty"`
	Currency          string     `json:"currency,omitempty"`
	FirstTradeDate    *time.Time `json:"first_trade_date,omitempty"`
	LastTradeDate     *time.Time `json:"last_trade_date,omitempty"`
	ExpirationDate    *time.Time `json:"expiration_date,omitempty"`
	TradingStatus     string     `json:"trading_status,omitempty"`
	MappingSource     string     `json:"mapping_source,omitempty"` // manual, root_hint, catalog_match
}

type TinkoffFuture struct {
	ExecutionSymbolRef
	Name        string `json:"name,omitempty"`
	ShortName   string `json:"short_name,omitempty"`
	BasicAsset  string `json:"basic_asset,omitempty"`
	BasicAssetSize Decimal `json:"basic_asset_size,omitempty"`
}

type Decimal struct {
	Value    string   `json:"value"` // exact decimal string for execution systems
	Approx   *float64 `json:"approx,omitempty"` // optional LLM/display helper
	Currency string   `json:"currency,omitempty"`
}

type OrderBookSnapshot struct {
	InstrumentUID string    `json:"instrument_uid"`
	Depth         int       `json:"depth"`
	Bid           *Decimal  `json:"bid,omitempty"`
	Ask           *Decimal  `json:"ask,omitempty"`
	Spread        *Decimal  `json:"spread,omitempty"`
	SpreadPct     *float64  `json:"spread_pct,omitempty"`
	AsOf          time.Time `json:"as_of"`
	Status        string    `json:"status"` // ok, empty, stale, error
	Source        string    `json:"source"` // tinkoff_orderbook
}

type FuturesMargin struct {
	InstrumentUID      string    `json:"instrument_uid"`
	InitialMargin      Decimal   `json:"initial_margin,omitempty"`
	MaintenanceMargin  Decimal   `json:"maintenance_margin,omitempty"`
	Currency           string    `json:"currency,omitempty"`
	AsOf               time.Time `json:"as_of"`
	Source             string    `json:"source"` // tinkoff_futures_margin
}

type InstrumentStatus struct {
	InstrumentUID string    `json:"instrument_uid"`
	TradingStatus string    `json:"trading_status"`
	BuyAvailable  bool      `json:"buy_available"`
	SellAvailable bool      `json:"sell_available"`
	APITradeAvailable bool   `json:"api_trade_available"`
	ExchangeOpen  bool      `json:"exchange_open"`
	AsOf          time.Time `json:"as_of"`
	Source        string    `json:"source"` // tinkoff_trading_status
}

type ResolverWarning struct {
	Code      string   `json:"code"`
	Severity  string   `json:"severity"` // info, warning, critical
	Message   string   `json:"message"`
	FieldPath string   `json:"field_path,omitempty"`
	SourceRefs []string `json:"source_refs,omitempty"`
}

type ResolverSource struct {
	ID          string    `json:"id"`
	Source     string    `json:"source"` // tradingview_mcp, resolver_db, tinkoff_instruments, tinkoff_orderbook
	AsOf        time.Time `json:"as_of"`
	Status      string    `json:"status"`
	Reliability string    `json:"reliability"`
}

type CandidateRequest struct {
	AnalysisSymbol AnalysisSymbolRef `json:"analysis_symbol"`
	AsOf           time.Time         `json:"as_of"`
	Limit          int               `json:"limit,omitempty"`
}

type CandidateResult struct {
	Success        bool                 `json:"success"`
	AnalysisSymbol AnalysisSymbolRef    `json:"analysis_symbol"`
	Candidates     []ExecutionSymbolRef `json:"candidates"`
	Warnings       []ResolverWarning    `json:"warnings,omitempty"`
}

type RefreshCatalogRequest struct {
	Provider string `json:"provider"` // tinkoff
	Force    bool   `json:"force,omitempty"`
}

type RefreshCatalogResult struct {
	Success      bool      `json:"success"`
	Provider     string    `json:"provider"`
	InstrumentCount int    `json:"instrument_count"`
	RefreshedAt  time.Time `json:"refreshed_at"`
}

type MappingKey struct {
	AnalysisProvider string `json:"analysis_provider"` // tradingview
	AnalysisSymbol   string `json:"analysis_symbol"`
	ExecutionProvider string `json:"execution_provider"` // tinkoff
}

type ManualMapping struct {
	MappingKey
	ExecutionInstrumentUID string     `json:"execution_instrument_uid"`
	FIGI                   string     `json:"figi,omitempty"`
	ValidFrom              time.Time  `json:"valid_from"`
	ValidUntil             *time.Time `json:"valid_until,omitempty"`
	Priority               int        `json:"priority"`
	Reason                 string     `json:"reason"`
	CreatedBy              string     `json:"created_by,omitempty"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

type ValidationResult struct {
	Success  bool              `json:"success"`
	Status   ResolutionStatus  `json:"status"`
	Warnings []ResolverWarning `json:"warnings,omitempty"`
}
```

Notes:

- `Decimal.Value` is the execution-safe value. `Approx` is optional and should
  be used only for display or LLM summaries.
- `OrderBookSnapshot` must come from Tinkoff for executable bid/ask/spread.
- `ExecutionSymbolRef` is resolved only when `InstrumentUID` is present and
  status is compatible with trading.

---

## SQLite Schema

This schema is proposed for the external HTS resolver store.

```sql
PRAGMA foreign_keys = ON;

CREATE TABLE analysis_symbols (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    provider TEXT NOT NULL,
    raw_symbol TEXT NOT NULL,
    exchange TEXT,
    ticker TEXT,
    root TEXT,
    roll_number INTEGER,
    is_continuous INTEGER NOT NULL DEFAULT 0,
    description TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider, raw_symbol)
);

CREATE TABLE tinkoff_futures (
    instrument_uid TEXT PRIMARY KEY,
    figi TEXT,
    asset_uid TEXT,
    position_uid TEXT,
    ticker TEXT NOT NULL,
    class_code TEXT,
    exchange TEXT,
    name TEXT,
    short_name TEXT,
    basic_asset TEXT,
    basic_asset_size TEXT,
    lot INTEGER,
    min_price_increment TEXT,
    currency TEXT,
    first_trade_date TEXT,
    last_trade_date TEXT,
    expiration_date TEXT,
    trading_status TEXT,
    buy_available INTEGER,
    sell_available INTEGER,
    api_trade_available INTEGER,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tinkoff_futures_ticker_class
    ON tinkoff_futures(ticker, class_code);

CREATE INDEX idx_tinkoff_futures_expiration
    ON tinkoff_futures(expiration_date);

CREATE INDEX idx_tinkoff_futures_basic_asset
    ON tinkoff_futures(basic_asset);

CREATE TABLE root_mapping_hints (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    analysis_provider TEXT NOT NULL,
    analysis_exchange TEXT,
    analysis_root TEXT NOT NULL,
    execution_provider TEXT NOT NULL,
    execution_class_code TEXT,
    execution_root_hint TEXT,
    execution_exchange TEXT,
    priority INTEGER NOT NULL DEFAULT 100,
    active INTEGER NOT NULL DEFAULT 1,
    notes TEXT,
    UNIQUE(
        analysis_provider,
        analysis_exchange,
        analysis_root,
        execution_provider,
        execution_class_code,
        execution_root_hint
    )
);

CREATE TABLE instrument_mappings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    analysis_symbol_id INTEGER NOT NULL REFERENCES analysis_symbols(id),
    execution_provider TEXT NOT NULL,
    execution_instrument_uid TEXT NOT NULL REFERENCES tinkoff_futures(instrument_uid),
    mapping_type TEXT NOT NULL,
    priority INTEGER NOT NULL DEFAULT 100,
    valid_from TEXT NOT NULL,
    valid_until TEXT,
    confidence REAL NOT NULL DEFAULT 1.0,
    reason TEXT,
    created_by TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    active INTEGER NOT NULL DEFAULT 1,
    UNIQUE(analysis_symbol_id, execution_provider, execution_instrument_uid, valid_from)
);

CREATE INDEX idx_instrument_mappings_active
    ON instrument_mappings(analysis_symbol_id, execution_provider, active, valid_from, valid_until);

CREATE TABLE resolver_decisions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    request_hash TEXT NOT NULL,
    analysis_symbol_id INTEGER NOT NULL REFERENCES analysis_symbols(id),
    execution_instrument_uid TEXT REFERENCES tinkoff_futures(instrument_uid),
    status TEXT NOT NULL,
    selected_by TEXT,
    confidence REAL,
    reason TEXT,
    warnings_json TEXT,
    source_trace_json TEXT,
    as_of TEXT NOT NULL,
    expires_at TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_resolver_decisions_request_hash
    ON resolver_decisions(request_hash, expires_at);

CREATE TABLE orderbook_snapshots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    instrument_uid TEXT NOT NULL REFERENCES tinkoff_futures(instrument_uid),
    depth INTEGER NOT NULL,
    bid TEXT,
    ask TEXT,
    spread TEXT,
    spread_pct REAL,
    status TEXT NOT NULL,
    as_of TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_orderbook_snapshots_uid_asof
    ON orderbook_snapshots(instrument_uid, as_of);

CREATE TABLE margin_snapshots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    instrument_uid TEXT NOT NULL REFERENCES tinkoff_futures(instrument_uid),
    initial_margin TEXT,
    maintenance_margin TEXT,
    currency TEXT,
    as_of TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_margin_snapshots_uid_asof
    ON margin_snapshots(instrument_uid, as_of);

CREATE TABLE mapping_audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    mapping_id INTEGER REFERENCES instrument_mappings(id),
    action TEXT NOT NULL,
    actor TEXT,
    before_json TEXT,
    after_json TEXT,
    reason TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

Seed example:

```sql
INSERT INTO root_mapping_hints (
    analysis_provider,
    analysis_exchange,
    analysis_root,
    execution_provider,
    execution_class_code,
    execution_root_hint,
    execution_exchange,
    priority,
    notes
) VALUES (
    'tradingview',
    'RUS',
    'NG',
    'tinkoff',
    'SPBFUT',
    'NG',
    'MOEX',
    10,
    'Example natural gas root hint. Concrete contract still must be resolved from Tinkoff catalog.'
);
```

---

## Proposed CLI Commands

These commands are proposed for the external HTS CLI, not for `tv`.

```text
hts instrument refresh --provider tinkoff [--force]
hts instrument resolve --analysis RUS:NG1! [--timeframe 1D] [--as-of 2026-04-27T00:00:00Z] [--require-tradable]
hts instrument candidates --analysis RUS:NG1! [--as-of 2026-04-27T00:00:00Z] [--limit 10]
hts instrument mapping get --analysis RUS:NG1!
hts instrument mapping set --analysis RUS:NG1! --uid <instrument_uid> --valid-until 2026-06-01 --reason "front contract roll"
hts instrument mapping validate --analysis RUS:NG1!
hts marketdata orderbook --uid <instrument_uid> [--depth 20]
hts instrument status --uid <instrument_uid>
hts instrument margin --uid <instrument_uid>
```

CLI behavior:

- JSON stdout by default for automation.
- Human diagnostics go to stderr.
- Non-zero exit when `status` is `unresolved`, `ambiguous`, `not_tradable`, or
  the provider is unavailable.
- `resolve` must not return a copied TradingView continuous symbol as the
  execution instrument.

---

## Proposed MCP Tools

These tools are proposed for the external HTS MCP, not for
`tradingview-mcp-go`.

| Tool | Purpose |
| --- | --- |
| `instrument_resolve` | Resolve TradingView analysis symbol to concrete Tinkoff execution instrument. |
| `instrument_candidates` | Return candidate Tinkoff futures for manual review. |
| `instrument_mapping_get` | Read current manual mapping. |
| `instrument_mapping_set` | Create/update manual mapping with audit metadata. |
| `instrument_mapping_validate` | Verify mapping still points to a tradable instrument. |
| `instrument_catalog_refresh` | Refresh Tinkoff futures catalog cache. |
| `marketdata_orderbook_get` | Get executable bid/ask/spread from Tinkoff orderbook. |
| `instrument_status_get` | Get trading status/session availability from Tinkoff. |
| `instrument_margin_get` | Get futures margin data from Tinkoff. |

Minimal `instrument_resolve` request:

```json
{
  "analysis_symbol": {
    "provider": "tradingview",
    "raw": "RUS:NG1!",
    "exchange": "RUS",
    "ticker": "NG1!",
    "root": "NG",
    "roll_number": 1,
    "is_continuous": true
  },
  "timeframe": "1D",
  "as_of": "2026-04-27T00:00:00Z",
  "orderbook_depth": 20,
  "require_tradable": true
}
```

Resolved response shape:

```json
{
  "success": true,
  "status": "resolved",
  "analysis_symbol": {
    "provider": "tradingview",
    "raw": "RUS:NG1!",
    "exchange": "RUS",
    "ticker": "NG1!",
    "root": "NG",
    "roll_number": 1,
    "is_continuous": true
  },
  "execution_symbol": {
    "provider": "tinkoff",
    "symbol": "SPBFUT:NGM6",
    "ticker": "NGM6",
    "class_code": "SPBFUT",
    "exchange": "MOEX",
    "figi": "example-figi",
    "instrument_uid": "example-instrument-uid",
    "lot": 1,
    "min_price_increment": { "value": "0.001", "approx": 0.001 },
    "currency": "rub",
    "expiration_date": "2026-06-18T00:00:00Z",
    "trading_status": "normal_trading",
    "mapping_source": "auto_roll"
  },
  "selected_by": "auto_roll",
  "confidence": 0.82,
  "orderbook": {
    "instrument_uid": "example-instrument-uid",
    "depth": 20,
    "bid": { "value": "3.124", "approx": 3.124 },
    "ask": { "value": "3.126", "approx": 3.126 },
    "spread": { "value": "0.002", "approx": 0.002 },
    "spread_pct": 0.064,
    "as_of": "2026-04-27T00:00:01Z",
    "status": "ok",
    "source": "tinkoff_orderbook"
  },
  "margin": {
    "instrument_uid": "example-instrument-uid",
    "initial_margin": { "value": "12500", "currency": "rub" },
    "maintenance_margin": { "value": "10000", "currency": "rub" },
    "currency": "rub",
    "as_of": "2026-04-27T00:00:01Z",
    "source": "tinkoff_futures_margin"
  },
  "instrument_status": {
    "instrument_uid": "example-instrument-uid",
    "trading_status": "normal_trading",
    "buy_available": true,
    "sell_available": true,
    "api_trade_available": true,
    "exchange_open": true,
    "as_of": "2026-04-27T00:00:01Z",
    "source": "tinkoff_trading_status"
  },
  "warnings": [
    {
      "code": "CONTINUOUS_CONTRACT_ANALYSIS_ONLY",
      "severity": "info",
      "message": "TradingView continuous contract is used only for analysis."
    }
  ],
  "recommended_action": "use_execution_symbol_for_marketdata_and_execution",
  "resolved_at": "2026-04-27T00:00:01Z",
  "expires_at": "2026-04-27T00:01:01Z"
}
```

---

## Fallback Behavior

When a symbol cannot be resolved, the resolver must fail closed.

### No candidates

Return:

```json
{
  "success": false,
  "status": "unresolved",
  "reason": "no Tinkoff futures candidates matched TradingView analysis root",
  "execution_symbol": null,
  "recommended_action": "manual_mapping_required",
  "warnings": [
    {
      "code": "EXECUTION_SYMBOL_UNRESOLVED",
      "severity": "critical",
      "message": "Analysis may continue, but execution data and orderbook are unavailable."
    }
  ]
}
```

HTS behavior:

- Allow analysis-only summary.
- Mark `execution_symbol.status="unresolved"`.
- Mark `risk_context.status="unavailable"` or `watch_only`.
- Do not calculate executable spread.
- Do not place orders.

### Ambiguous candidates

Return `status="ambiguous"` with candidates and require manual mapping. Do not
select by ticker similarity alone when expiration or root identity is unclear.

### Candidate not tradable

Return `status="not_tradable"` when the instrument exists but status, session,
API trade availability, or bid/ask makes it unsuitable for execution.

### Tinkoff provider unavailable

Return `status="stale"` only if a non-expired cached resolver decision exists.
Otherwise return `status="unresolved"` and `recommended_action="retry_provider"`.

### TradingView symbol is not continuous

For non-continuous symbols, the resolver may still map by manual mapping or
catalog hints, but it must preserve `analysis_symbol.is_continuous=false` and
must not assume the TradingView ticker is executable.

---

## Warning Codes

- `CONTINUOUS_CONTRACT_ANALYSIS_ONLY`: TradingView continuous symbol is not an
  execution symbol.
- `EXECUTION_SYMBOL_UNRESOLVED`: no concrete Tinkoff instrument selected.
- `EXECUTION_SYMBOL_AMBIGUOUS`: multiple plausible candidates.
- `MANUAL_MAPPING_REQUIRED`: resolver needs operator confirmation.
- `MANUAL_MAPPING_EXPIRED`: stored mapping is outside validity range.
- `TINKOFF_CATALOG_STALE`: instrument catalog cache is stale.
- `TINKOFF_PROVIDER_UNAVAILABLE`: Tinkoff API unavailable.
- `ORDERBOOK_UNAVAILABLE`: bid/ask/spread cannot be obtained.
- `ORDERBOOK_EMPTY`: orderbook returned no usable bid/ask.
- `SPREAD_TOO_WIDE`: spread exceeds configured threshold.
- `INSTRUMENT_NOT_TRADABLE`: status or session blocks execution.
- `MARGIN_UNAVAILABLE`: futures margin data missing.
- `EXPIRATION_NEAR`: instrument is close to expiration.
- `ROLL_MAPPING_LOW_CONFIDENCE`: auto roll selection has weak evidence.

---

## Integration With Market Summary

The resolver feeds the following fields in `HTS Market Summary Contract`:

- `execution_symbol`: from `ResolutionResult.ExecutionSymbol`.
- `current_price.bid`, `current_price.ask`, `current_price.mid`: from Tinkoff
  orderbook.
- `risk_context.bid`, `risk_context.ask`, `risk_context.spread`,
  `risk_context.spread_pct`: from Tinkoff orderbook.
- `risk_context.lot`, `risk_context.min_price_increment`: from Tinkoff
  instrument.
- expiration, margin, and trading status: from Tinkoff instrument/margin/status
  calls.
- `warnings`: copied from resolver warning codes.
- `source_trace`: include resolver DB, Tinkoff catalog, Tinkoff orderbook,
  margin, and status source entries.

TradingView remains the analysis source for chart state, OHLCV, studies,
screenshots, Pine output, and visual context.
