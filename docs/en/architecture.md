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

The project intentionally maintains **1:1 behavioral compatibility** with the original Node.js implementation:

| Aspect | Node.js | Go port |
| --- | --- | --- |
| MCP tools | 78 | 78 (100% name match) |
| CLI groups | 15 | 15+ |
| Argument JSON schemas | original | identical |
| Response JSON structure | `{success, ...}` | identical |
| CDP endpoint | `localhost:9222` | identical |
| JS expressions | `chrome-remote-interface` | identical |
| Platforms | Win / macOS / Linux | Win / macOS / Linux |
| Windows Store | yes | yes (Get-AppxPackage) |

Full compatibility matrix: [docs/dev/COMPATIBILITY_MATRIX.md](../dev/COMPATIBILITY_MATRIX.md)

---

## Disclaimer

This tool communicates **only** with your locally running TradingView Desktop instance via the Chrome DevTools Protocol at `localhost:9222`.

- Does not connect to TradingView servers.
- Does not execute real trades.
- Does not collect or transmit market data outside your local machine.
- Does not bypass any TradingView paid features.
- Not affiliated with TradingView Inc. or Anthropic.
