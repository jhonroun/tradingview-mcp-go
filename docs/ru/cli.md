# CLI-команды и скрипты

> [← Назад к документации](README.md)

---

## CLI (`tv`)

```bash
# Диагностика
tv status                          # Проверить CDP
tv launch [--port N] [--no-kill] [--tv-path PATH]
tv doctor                          # Диагностика установки и CDP
tv discover                        # Доступные API пути
tv ui-state                        # Открытые панели

# График
tv chart-state
tv set-symbol SYMBOL
tv set-timeframe TF                # 1 5 15 60 D W M
tv set-type TYPE                   # Candles HeikinAshi Line Area ...

# Данные
tv quote [SYMBOL]
tv ohlcv [--count N] [--summary]
tv screenshot [--region full|chart|strategy_tester] [--filename F]

# Символы
tv symbol-info
tv symbol-search QUERY [--type T] [--exchange E]

# Индикаторы
tv indicator-toggle ENTITY_ID [--visible=true|false]

# Pine Script
tv pine get
tv pine set "SOURCE"
tv pine compile
tv pine smart-compile
tv pine raw-compile
tv pine errors
tv pine console
tv pine save
tv pine new [indicator|strategy|library]
tv pine open NAME
tv pine list
tv pine analyze "SOURCE"
tv pine check "SOURCE"

# Рисование
tv draw shape --time T --price P [--time2 T2 --price2 P2] [--text TEXT]
tv draw list
tv draw get ENTITY_ID
tv draw remove ENTITY_ID
tv draw clear

# Панели
tv pane list
tv pane set-layout LAYOUT
tv pane focus INDEX
tv pane set-symbol SYMBOL

# Вкладки
tv tab list
tv tab new
tv tab close
tv tab switch ID

# Replay
tv replay start [--date YYYY-MM-DD]
tv replay step
tv replay stop
tv replay status
tv replay autoplay [--speed MS]
tv replay trade buy|sell|close

# Алерты и Watchlist
tv alert list
tv alert create --price P [--message MSG]
tv alert delete [--all]
tv watchlist get
tv watchlist add SYMBOL

# UI
tv ui click --by aria-label --value "..."
tv ui open-panel PANEL [--action open|close|toggle]
tv ui fullscreen
tv ui keyboard KEY [--modifiers ctrl,shift]
tv ui type TEXT
tv ui hover --by aria-label --value "..."
tv ui scroll up|down [--amount N]
tv ui mouse X Y [--button right] [--double]
tv ui find QUERY [--strategy text|css|aria-label]
tv ui eval "JS_EXPRESSION"

# Лейауты
tv layout list
tv layout switch NAME

# Пакетная обработка
tv batch --symbols SYM1,SYM2 --action screenshot|get_ohlcv|get_strategy_results \
         [--timeframes TF1,TF2] [--delay MS] [--count N]

# JSONL-стримы (Ctrl+C для остановки)
tv stream quote    [--interval MS]
tv stream bars     [--interval MS]
tv stream values   [--interval MS]
tv stream lines    [--interval MS] [--filter STUDY]
tv stream labels   [--interval MS] [--filter STUDY]
tv stream tables   [--interval MS] [--filter STUDY]
tv stream all      [--interval MS]
```

---

## Скрипты

### Запуск TradingView

| Скрипт | Платформа | Описание |
| --- | --- | --- |
| `scripts/launch_tv_debug.bat` | Windows | Запуск TradingView с CDP |
| `scripts/launch_tv_debug.vbs` | Windows | Тихий запуск (без окна cmd) |
| `scripts/launch_tv_debug_mac.sh` | macOS | Запуск TradingView с CDP |
| `scripts/launch_tv_debug_linux.sh` | Linux | Запуск TradingView с CDP |

### Pine Script

| Скрипт | Описание |
| --- | --- |
| `scripts/pine_pull.sh` / `.bat` | Извлечь исходник из редактора → `scripts/current.pine` |
| `scripts/pine_push.sh` / `.bat` | Загрузить `scripts/current.pine` в редактор + скомпилировать |

### Сборка и установка

| Скрипт | Описание |
| --- | --- |
| `scripts/build.sh` | Сборка для текущей платформы → `bin/` |
| `scripts/build.bat` | То же для Windows |
| `scripts/install.sh` | Установка в `/usr/local/bin` (или `PREFIX`) |
| `scripts/install.bat` | Установка в `%SystemRoot%\System32` |
| `scripts/bootstrap.sh` | Curl-pipe установщик (скачивает бинарники с GitHub) |
| `scripts/bootstrap.ps1` | PowerShell установщик для Windows |
| `scripts/configure-mcp.sh` | Настройка MCP-конфига для указанного клиента |
| `scripts/configure-mcp.ps1` | То же для Windows |
| `scripts/package.sh` | Создать release-архивы для всех платформ |
