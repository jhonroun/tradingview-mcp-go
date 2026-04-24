# COMPATIBILITY_MATRIX.md — матрица совместимости Node.js → Go

## Назначение

Файл фиксирует соответствие исходной Node.js-реализации и Go-порта.

Перед закрытием каждого этапа нужно заполнить соответствующие строки.

## Общие контракты

| Компонент | Node.js | Go | Статус |
|---|---|---|---|
| MCP server | `src/server.js` | `cmd/tvmcp`, `internal/mcp` | pending |
| CLI | `src/cli/index.js` | `cmd/tv`, `internal/cli` | pending |
| CDP connection | `src/connection.js` | `internal/cdp` | pending |
| Runtime.evaluate | `chrome-remote-interface` | Go CDP client | pending |
| Tools registry | Node SDK | Go registry | pending |
| Streaming | CLI JSONL | Go JSONL | pending |

## Tool groups

| Группа | Node.js | Go | Статус |
|---|---|---|---|
| Health | available | pending | pending |
| Chart reading | available | pending | pending |
| Quote/OHLCV | available | pending | pending |
| Pine Script | available | pending | pending |
| Drawing | available | pending | pending |
| Alerts | available | pending | pending |
| Watchlist | available | pending | pending |
| Indicator | available | pending | pending |
| Layout | available | pending | pending |
| Pane | available | pending | pending |
| Tab | available | pending | pending |
| Replay | available | pending | pending |
| UI automation | available | pending | pending |
| Screenshot | available | pending | pending |
| Discover/UI state/range/scroll | available | pending | pending |

## CLI commands

| Command | Go status | Notes |
|---|---|---|
| `tv status` | pending | first command |
| `tv launch` | pending | platform-specific |
| `tv state` | pending |  |
| `tv symbol` | pending |  |
| `tv timeframe` | pending |  |
| `tv type` | pending |  |
| `tv info` | pending |  |
| `tv search` | pending |  |
| `tv quote` | pending |  |
| `tv ohlcv` | pending |  |
| `tv values` | pending |  |
| `tv data` | pending | group |
| `tv pine` | pending | group |
| `tv draw` | pending | group |
| `tv alert` | pending | group |
| `tv watchlist` | pending | group |
| `tv indicator` | pending | group |
| `tv layout` | pending | group |
| `tv pane` | pending | group |
| `tv tab` | pending | group |
| `tv replay` | pending | group |
| `tv stream` | pending | group |
| `tv ui` | pending | group |
| `tv screenshot` | pending |  |
| `tv discover` | pending |  |
| `tv ui-state` | pending |  |
| `tv range` | pending |  |
| `tv scroll` | pending |  |

## Проверка строки

Статусы:

- `pending` — не начато;
- `in-progress` — переносится;
- `compatible` — совпадает с Node.js;
- `partial` — частично совпадает;
- `changed` — отличается, нужно обоснование;
- `blocked` — заблокировано внешним фактором.


## Windows discovery / launcher

| Компонент | Node.js | Go | Статус | Примечание |
|---|---|---|---|---|
| Standard Windows install | launch script / auto-detect | `internal/discovery/windows.go` | pending | `%LOCALAPPDATA%`, `%PROGRAMFILES%` |
| Microsoft Store install | часто проблемно | `Get-AppxPackage` + WindowsApps fallback | pending | `C:\Program Files\WindowsApps\TradingView.Desktop_*` |
| Manual override | частично | env/flag/config | pending | `TRADINGVIEW_PATH`, `--tv-path` |
| Diagnostics | частично | `tv doctor windows` | pending | обязательный troubleshooting tool |

## Skills / scripts / utilities

| Группа | Node.js | Go | Статус |
|---|---|---|---|
| Launch scripts | `scripts/launch_tv_debug_*` | launcher + wrapper scripts | pending |
| CLI utilities | `tv ...` | `cmd/tv` | pending |
| Health tools | `tv_health_check` | MCP + CLI | pending |
| Workflow skills | исходный repo/fork dependent | Go workflows | pending |
