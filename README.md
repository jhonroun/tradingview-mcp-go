# tradingview-mcp-go

AI Go port of [tradesdontlie/tradingview-mcp](https://github.com/tradesdontlie/tradingview-mcp).

Connect any MCP client (Claude Code, Cursor, Cline…) to a live **TradingView Desktop** chart via the Chrome DevTools Protocol.

> Not affiliated with TradingView Inc. or Anthropic.  
> Ensure your usage complies with TradingView's Terms of Use.

---

## Documentation

| Language | Overview | Install | Tools | CLI | Agents & Skills | Architecture |
| --- | --- | --- | --- | --- | --- | --- |
| Русский | [README](docs/ru/README.md) | [install.md](docs/ru/install.md) | [tools.md](docs/ru/tools.md) | [cli.md](docs/ru/cli.md) | [agents-skills.md](docs/ru/agents-skills.md) | [architecture.md](docs/ru/architecture.md) |
| English | [README](docs/en/README.md) | [install.md](docs/en/install.md) | [tools.md](docs/en/tools.md) | [cli.md](docs/en/cli.md) | [agents-skills.md](docs/en/agents-skills.md) | [architecture.md](docs/en/architecture.md) |

Development process docs (porting history, compatibility matrix): [docs/dev/](docs/dev/README.md)

---

## Quick Start

### One-line install (no cloning required)

#### Linux / macOS

```bash
# Install only
curl -fsSL https://raw.githubusercontent.com/jhonroun/tradingview-mcp-go/main/scripts/bootstrap.sh | bash

# Install + auto-configure Claude Code
curl -fsSL https://raw.githubusercontent.com/jhonroun/tradingview-mcp-go/main/scripts/bootstrap.sh | CLIENT=claude bash

# Install + auto-configure Cursor
curl -fsSL https://raw.githubusercontent.com/jhonroun/tradingview-mcp-go/main/scripts/bootstrap.sh | CLIENT=cursor bash
```

#### Windows (PowerShell)

```powershell
# Install only
iwr -useb https://raw.githubusercontent.com/jhonroun/tradingview-mcp-go/main/scripts/bootstrap.ps1 | iex

# Install + auto-configure Claude Code
.\bootstrap.ps1 -Client claude

# Install + auto-configure Cursor with custom prefix
.\bootstrap.ps1 -Client cursor -Prefix "C:\tools\tvmcp"
```

Supported clients: `claude` · `cursor` · `cline` · `windsurf` · `continue` · `codex` · `gemini`

### Build from source

```bash
# 1. Build
bash scripts/build.sh          # Linux / macOS  →  bin/tvmcp  bin/tv
scripts\build.bat              # Windows        →  bin\tvmcp.exe  bin\tv.exe
# or:  make build

# 2. Start TradingView with CDP
tv launch
# or:  scripts\launch_tv_debug.bat  (Windows)

# 3. Add to your MCP client config
# { "mcpServers": { "tradingview": { "command": "/path/to/tvmcp" } } }

# 4. Verify
tv status
```

---

## What it does

78 MCP tools + CLI (`tv`) for:

- **Chart**: read state, set symbol / timeframe / type, manage indicators
- **Data**: OHLCV, quotes, strategy results, equity curve, order book
- **Pine Script**: read / write / compile / analyze source, list saved scripts
- **Drawing**: create / list / remove shapes and annotations
- **Alerts & Watchlist**: create, list, delete alerts; manage watchlist
- **Replay**: step through history, take practice trades, track P&L
- **UI automation**: click, scroll, keyboard, mouse, find elements, open panels
- **Streaming**: real-time JSONL output for quotes, bars, indicators, Pine objects
- **Batch**: iterate symbols × timeframes with screenshot / OHLCV / strategy actions

---

## Supported MCP Clients

`tvmcp` speaks standard MCP (JSON-RPC 2.0 over stdio) and works with any compatible client:

**Claude Code** · **Cursor** · **Cline** · **Continue** · **Windsurf** · any client implementing MCP stdio

---

## Project History

Originally a Node.js project. Ported to Go in April 2026 by [jhonroun](https://github.com/jhonroun) with direct assistance from **Claude Code** (Anthropic's AI developer CLI).

The port preserves 1:1 behavioral compatibility: same 78 tool names, same argument schemas, same JSON response structure.

Full story: [docs/ru/README.md → История создания](docs/ru/README.md#история-создания) · [docs/en/README.md → Origin Story](docs/en/README.md#origin-story)

---

## Repository Layout

```text
cmd/tvmcp/        MCP stdio server binary
cmd/tv/           CLI binary
internal/         Go packages (cdp, mcp, tools, stream, ...)
agents/           Agent files in native format per client
  cursor/ cline/ windsurf/ continue/ codex/ gemini/
prompts/          Universal system prompt (source of truth)
skills/           Workflow scenarios (chart-analysis, pine-develop, ...)
scripts/          Build, install, launch, configure, and pine helper scripts
docs/
  ru/             Russian documentation (6 files)
  en/             English documentation (6 files)
  dev/            Development process docs (porting history, plans, matrix)
Makefile          make build / build-all / install / test / release
CHANGELOG.md      Change history
```

---

## Disclaimer

Communicates **only** with your locally running TradingView Desktop via CDP at `localhost:9222`. Does not execute real trades, connect to TradingView servers, or transmit market data outside your machine.
