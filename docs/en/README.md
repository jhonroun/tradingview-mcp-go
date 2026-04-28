# tradingview-mcp-go — Documentation (EN)

> [Документация на русском](../ru/README.md) | [Project root](../../README.md)

AI Go port of [tradesdontlie/tradingview-mcp](https://github.com/tradesdontlie/tradingview-mcp).

Connects any MCP client to a live **TradingView Desktop** chart via the Chrome DevTools Protocol.

> Not affiliated with TradingView Inc. or Anthropic.  
> Ensure your usage complies with TradingView's Terms of Use.

---

## Navigation

| Section | File |
| --- | --- |
| About, origin story, requirements | this page |
| Install, CDP launch, MCP config, providers | [install.md](install.md) |
| MCP tools (85 current Go tools) | [tools.md](tools.md) |
| CLI commands and scripts | [cli.md](cli.md) |
| Skills and agents | [agents-skills.md](agents-skills.md) |
| Architecture, compatibility, disclaimer | [architecture.md](architecture.md) |

---

## About

`tradingview-mcp-go` is an MCP server and CLI utility that allows AI assistants (Claude Code, Cursor, Cline, and other MCP clients) to interact with a running TradingView Desktop:

- read and control charts (symbol, timeframe, indicators);
- access real-time market data (OHLCV, quotes, strategy results);
- work with Pine Script (read, write, compile, analyze);
- draw shapes, manage alerts, watchlists, panes, and tabs;
- run manual backtests in Replay mode;
- automate TradingView UI programmatically;
- stream data as JSONL.

All communication is **local only**, via Chrome DevTools Protocol at `localhost:9222`. No data is sent to external servers.

---

## Origin Story

### Original Project

The repository [`tradesdontlie/tradingview-mcp`](https://github.com/tradesdontlie/tradingview-mcp) is a **Node.js** MCP server implementation providing 78 tools for interacting with TradingView Desktop via CDP.

### Porting to Go with Claude Code

In April 2026, the entire Node.js project was ported to **Go** with 1:1 behavioral compatibility.

The porting process was performed with the direct assistance of **Claude Code** (Anthropic's CLI for developers):

- inventoried the Node.js source and built the compatibility matrix;
- implemented each module sequentially (MCP, CDP, tools, CLI, stream);
- wrote unit tests after each phase and verified `go test ./...`;
- maintained `CHANGELOG.md` and `TODO.md` throughout the process;
- ensured 1:1 compatibility in tool names, argument schemas, and JSON response structure.

Initial parity result: **78/78 MCP tools, 83+ CLI commands, 78 unit tests** — in a single session.

Current stabilized Go registry: **85 MCP tools**. The additional tools are Go extensions for study-model history, strategy orders, Pine restore safety, and LLM-ready aggregate context.

**Outcome:** a fully functional Go binary with zero dependency on Node.js, `npm`, or `chrome-remote-interface`.

### Current Reliability Boundaries

- TradingView internal paths are undocumented. Use `tv discover` compatibility probes after TradingView Desktop updates.
- Strategy equity from `data_get_equity` is loaded-bars-only (`coverage: loaded_chart_bars`) unless TradingView has actually loaded the full range.
- Derived equity is conditional and not native Strategy Tester equity.
- Full native bar-by-bar Strategy Tester equity is not implemented unless TradingView exposes a stable report field.

---

## Requirements

| Component | Version |
| --- | --- |
| Go | 1.21+ |
| TradingView Desktop | Windows / macOS / Linux |
| TradingView must run with | `--remote-debugging-port=9222` |

No Node.js dependency. Single Go dependency: `github.com/gorilla/websocket`.
