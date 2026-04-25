# README_PORTING_START.md

## Быстрый старт портирования

```bash
git clone https://github.com/tradesdontlie/tradingview-mcp.git tradingview-mcp-node
mkdir tradingview-mcp-go
cd tradingview-mcp-go
go mod init github.com/YOURNAME/tradingview-mcp-go
```

## Первый коммит Go-порта

Создать:

```text
cmd/tvmcp/main.go
cmd/tv/main.go
internal/mcp/
internal/cdp/
internal/tools/health/
PLAN.md
TODO.md
CHANGELOG.md
PORTING_GUIDE.md
COMPATIBILITY_MATRIX.md
```

## Первый ожидаемый результат

```bash
go test ./...
go run ./cmd/tv status
go run ./cmd/tvmcp
```

`tv status` должен вернуть JSON, даже если TradingView не запущен:

```json
{
  "ok": false,
  "error": "CDP connection failed: localhost:9222 is not available"
}
```

Когда TradingView запущен с debug port:

```json
{
  "ok": true,
  "cdp": {
    "host": "localhost",
    "port": 9222
  },
  "target": {
    "type": "page",
    "url": "https://www.tradingview.com/chart/..."
  }
}
```

## MCP config для Claude Code

После сборки:

```json
{
  "mcpServers": {
    "tradingview-go": {
      "command": "/absolute/path/to/tvmcp",
      "args": []
    }
  }
}
```

## Важное ограничение

До прохождения P4 не начинать массовое портирование tools.

Сначала нужно доказать, что Go MCP server + Go CDP client работают end-to-end.
