# Development Documentation

Internal documents from the porting process (Node.js → Go).

Current note: the original parity baseline was 78 Node-compatible tools. The
stabilized Go registry currently exposes 85 tools after adding study history,
orders, Pine restore safety, and aggregate context helpers.

| Document | Description |
| --- | --- |
| [PLAN.md](PLAN.md) | Strategic porting plan |
| [TODO.md](TODO.md) | Task checklist by phase |
| [PORTING_GUIDE.md](PORTING_GUIDE.md) | Node.js → Go mapping rules |
| [PORTING_NOTES.md](PORTING_NOTES.md) | Facts found while reading Node.js source |
| [COMPATIBILITY_MATRIX.md](COMPATIBILITY_MATRIX.md) | Tool-by-tool Node.js → Go comparison |
| [FINAL_AUDIT_REPORT.md](FINAL_AUDIT_REPORT.md) | Final HTS readiness audit and decision |
| [HTS_MARKET_SUMMARY_CONTRACT.md](HTS_MARKET_SUMMARY_CONTRACT.md) | Draft source-aware market summary contract for TradingView MCP -> external HTS MCP -> LLM |
| [LLM_MARKET_CONTEXT_CONTRACT.md](LLM_MARKET_CONTEXT_CONTRACT.md) | Draft compact LLM context contract: single instrument, scan, diff, trade review, and LLM response schema |
| [INSTRUMENT_RESOLVER_CONTRACT.md](INSTRUMENT_RESOLVER_CONTRACT.md) | Draft resolver contract for TradingView analysis symbols -> Tinkoff execution instruments |
| [ORCHESTRATION_AND_LIMITS.md](ORCHESTRATION_AND_LIMITS.md) | How the porting was orchestrated within Claude Pro limits |
| [AGENTS.md](AGENTS.md) | Agent roles during the porting process |
| [README_PORTING_START.md](README_PORTING_START.md) | Quick-start guide used at the beginning of the port |
