# Installation and Setup

> [← Back to docs](README.md)

---

## Install binaries

### Option 0: Bootstrap (recommended, no cloning required)

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

# Install + configure Claude Code
.\bootstrap.ps1 -Client claude

# Install + configure Cursor with custom prefix
.\bootstrap.ps1 -Client cursor -Prefix "C:\tools\tvmcp"
```

Supported clients: `claude` · `cursor` · `cline` · `windsurf` · `continue` · `codex` · `gemini`

### Option 1: go install

```bash
go install github.com/jhonroun/tradingview-mcp-go/cmd/tvmcp@latest
go install github.com/jhonroun/tradingview-mcp-go/cmd/tv@latest
```

### Option 2: Build from source

```bash
git clone https://github.com/jhonroun/tradingview-mcp-go
cd tradingview-mcp-go

# Linux / macOS
bash scripts/build.sh
# Outputs: bin/tvmcp  bin/tv

# Windows
scripts\build.bat
# Outputs: bin\tvmcp.exe  bin\tv.exe
```

### Option 3: Make

```bash
make build        # current platform → bin/
make build-all    # all platforms: windows/linux/darwin × amd64/arm64
make install      # go install into $GOPATH/bin
make test         # go test ./...
make release      # build-all + ZIP/tar.gz archives
```

### Install to system PATH

```bash
# Linux / macOS
sudo bash scripts/install.sh
# or manually:
sudo cp bin/tvmcp bin/tv /usr/local/bin/

# Windows (as Administrator)
scripts\install.bat
# or manually:
copy bin\tvmcp.exe %SystemRoot%\System32\
copy bin\tv.exe    %SystemRoot%\System32\
```

---

## Starting TradingView with CDP

TradingView Desktop must be started with `--remote-debugging-port=9222`.

### Option 1: Auto-launch via CLI

```bash
tv launch
# or with an explicit path:
tv launch --tv-path="C:\Users\you\AppData\Local\TradingView\TradingView.exe"
```

### Option 2: Launch scripts

```bash
# Windows
scripts\launch_tv_debug.bat

# macOS
bash scripts/launch_tv_debug_mac.sh

# Linux
bash scripts/launch_tv_debug_linux.sh
```

### Option 3: Manual

```bash
# Windows
"C:\Users\<you>\AppData\Local\TradingView\TradingView.exe" --remote-debugging-port=9222

# macOS
/Applications/TradingView.app/Contents/MacOS/TradingView --remote-debugging-port=9222

# Linux
tradingview --remote-debugging-port=9222
```

### Verify connection

```bash
tv status
tv doctor
```

---

## MCP Configuration

Add `tvmcp` to your MCP client configuration.

### Claude Code (`~/.claude.json`)

```json
{
  "mcpServers": {
    "tradingview": {
      "command": "/usr/local/bin/tvmcp"
    }
  }
}
```

Windows:

```json
{
  "mcpServers": {
    "tradingview": {
      "command": "C:\\Users\\you\\AppData\\Local\\tvmcp\\tvmcp.exe"
    }
  }
}
```

### Cursor (`~/.cursor/mcp.json` or `%APPDATA%\Cursor\User\mcp.json`)

```json
{
  "mcpServers": {
    "tradingview": {
      "command": "/usr/local/bin/tvmcp"
    }
  }
}
```

### Cline, Continue, Windsurf

Use the same `mcpServers` format — it is the standard MCP protocol configuration.  
Config file paths for each client: `scripts/configure-mcp.sh --list`

### Auto-configure

```bash
# Linux / macOS
bash scripts/configure-mcp.sh --client claude
bash scripts/configure-mcp.sh --client cursor

# Windows
.\scripts\configure-mcp.ps1 -Client claude
.\scripts\configure-mcp.ps1 -Client cursor
```

### Verify after setup

```bash
tv status        # check CDP connection
tv discover      # check available API paths
```

`tv discover` also returns `compatibility_probes` for undocumented TradingView internals. After TradingView Desktop updates, verify those probes before relying on study model values, strategy reports, or equity extraction. Strategy equity remains `coverage: loaded_chart_bars`; derived equity is conditional and not native Strategy Tester equity.

---

## Supported AI Providers

`tvmcp` is a standard MCP server over **stdio** (JSON-RPC 2.0). It is **not tied** to Claude Code or Anthropic.

| Client | MCP Support |
| --- | --- |
| **Claude Code** (Anthropic CLI) | Yes — primary development platform |
| **Cursor** | Yes — via MCP settings |
| **Cline** (VS Code extension) | Yes |
| **Continue** (VS Code extension) | Yes |
| **Windsurf** | Yes |
| **OpenAI Codex** | Only if it implements MCP stdio |
| Any custom client | Yes, if it supports JSON-RPC 2.0 / stdio MCP |

The server makes no network calls to AI providers — it only exposes tools.
