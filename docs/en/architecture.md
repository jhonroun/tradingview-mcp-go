# Architecture, Compatibility, and Disclaimer

> [← Back to docs](README.md)

---

## Architecture

```
AI Client (Claude Code, Cursor, Cline...)
    │
    │  MCP stdio (JSON-RPC 2.0)
    ▼
tvmcp (MCP server)
    │
    │  Chrome DevTools Protocol (WebSocket)
    ▼
TradingView Desktop  ←→  localhost:9222
```

### Package structure

```
cmd/tvmcp/        MCP stdio server
cmd/tv/           CLI utility
internal/
  cdp/            WebSocket CDP client (Runtime, Page, DOM, Input)
  mcp/            JSON-RPC 2.0 server + tool registry
  cli/            CLI command dispatcher
  stream/         JSONL poll-and-dedup streaming engine
  tradingview/    JS expression constants + SafeString
  discovery/      TradingView install discovery (Win Store / AppData / macOS / Linux)
  launcher/       Launch TradingView with --remote-debugging-port
  tools/
    health/       Health check and launch
    chart/        Chart state and control
    data/         OHLCV, quotes, indicators, strategies
    capture/      Screenshot
    indicators/   Indicator inputs and visibility
    pine/         Pine Script operations
    drawing/      Shape drawing
    alerts/       Alerts and watchlist
    replay/       Replay trading
    pane/         Pane management
    tab/          Tab management
    ui/           UI automation and layouts
    batch/        Batch operations
```

---

## Compatibility

The project intentionally preserves the original **78-tool Node.js parity baseline** and adds Go-only stabilization helpers.

| Aspect | Node.js | Go port |
| --- | --- | --- |
| MCP tools | 78 | 85 current Go tools: 78 parity tools + 7 extensions |
| CLI groups | 15 | 15+ |
| Argument JSON schemas | original | identical |
| Response JSON structure | `{success, ...}` | identical |
| CDP endpoint | `localhost:9222` | identical |
| JS expressions | `chrome-remote-interface` | identical |
| Platforms | Win / macOS / Linux | Win / macOS / Linux |
| Windows Store | yes | yes (Get-AppxPackage) |

Full compatibility matrix: [docs/dev/COMPATIBILITY_MATRIX.md](../dev/COMPATIBILITY_MATRIX.md)

Go-only extensions include `data_get_indicator_history`, `data_get_orders`, `pine_restore_source`, and four aggregate LLM/context tools. Data tools expose `source`, `reliability`, `coverage`, `status`, and `reliableForTradingLogic` where needed because several TradingView paths are undocumented internals.

`tv_discover` is the compatibility probe entry point for those internals. It preserves the legacy `paths` object and adds `compatibility_probes` with `compatible`, `available`, `status`, `stability`, and `reliability`. Run it after TradingView Desktop updates before trusting study model, backtesting API, or strategy equity plot workflows.

Strategy equity is intentionally modeled as loaded chart data, not as a full native Strategy Tester export. The only reliable runtime path is an explicit Pine `Strategy Equity` plot read from loaded bars; derived equity remains conditional and full native bar-by-bar equity is out of scope until TradingView exposes a stable report field.

---

## Disclaimer

This tool communicates **only** with your locally running TradingView Desktop instance via the Chrome DevTools Protocol at `localhost:9222`.

- Does not connect to TradingView servers.
- Does not execute real trades.
- Does not collect or transmit market data outside your local machine.
- Does not bypass any TradingView paid features.
- Not affiliated with TradingView Inc. or Anthropic.
