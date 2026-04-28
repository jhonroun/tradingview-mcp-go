# Установка и настройка

> [← Назад к документации](README.md)

---

## Установка бинарников

### Способ 0: Bootstrap (рекомендуется, без клонирования)

**Linux / macOS**

```bash
# Установка
curl -fsSL https://raw.githubusercontent.com/jhonroun/tradingview-mcp-go/main/scripts/bootstrap.sh | bash

# Установка + автонастройка MCP для Claude Code
curl -fsSL https://raw.githubusercontent.com/jhonroun/tradingview-mcp-go/main/scripts/bootstrap.sh | CLIENT=claude bash

# Установка + автонастройка для Cursor
curl -fsSL https://raw.githubusercontent.com/jhonroun/tradingview-mcp-go/main/scripts/bootstrap.sh | CLIENT=cursor bash
```

**Windows (PowerShell)**

```powershell
# Установка
iwr -useb https://raw.githubusercontent.com/jhonroun/tradingview-mcp-go/main/scripts/bootstrap.ps1 | iex

# Установка + настройка Claude Code
.\bootstrap.ps1 -Client claude

# Установка + настройка Cursor с кастомным путём
.\bootstrap.ps1 -Client cursor -Prefix "C:\tools\tvmcp"
```

Поддерживаемые клиенты: `claude` · `cursor` · `cline` · `windsurf` · `continue` · `codex` · `gemini`

### Способ 1: go install

```bash
go install github.com/jhonroun/tradingview-mcp-go/cmd/tvmcp@latest
go install github.com/jhonroun/tradingview-mcp-go/cmd/tv@latest
```

### Способ 2: сборка из исходников

```bash
git clone https://github.com/jhonroun/tradingview-mcp-go
cd tradingview-mcp-go

# Linux / macOS
bash scripts/build.sh
# Бинарники: bin/tvmcp  bin/tv

# Windows
scripts\build.bat
# Бинарники: bin\tvmcp.exe  bin\tv.exe
```

### Способ 3: Make

```bash
make build        # текущая платформа → bin/
make build-all    # все платформы: windows/linux/darwin × amd64/arm64
make install      # go install в $GOPATH/bin
make test         # go test ./...
make release      # build-all + ZIP/tar.gz архивы
```

### Установка в системный PATH

```bash
# Linux / macOS
sudo bash scripts/install.sh
# или вручную
sudo cp bin/tvmcp bin/tv /usr/local/bin/

# Windows (от Администратора)
scripts\install.bat
# или вручную
copy bin\tvmcp.exe %SystemRoot%\System32\
copy bin\tv.exe    %SystemRoot%\System32\
```

---

## Запуск TradingView с CDP

TradingView Desktop должен быть запущен с флагом `--remote-debugging-port=9222`.

### Способ 1: Автоматически через CLI

```bash
tv launch
# или с явным путём:
tv launch --tv-path="C:\Users\you\AppData\Local\TradingView\TradingView.exe"
```

### Способ 2: Скрипты запуска

```bash
# Windows
scripts\launch_tv_debug.bat

# macOS
bash scripts/launch_tv_debug_mac.sh

# Linux
bash scripts/launch_tv_debug_linux.sh
```

### Способ 3: Вручную

```bash
# Windows
"C:\Users\<you>\AppData\Local\TradingView\TradingView.exe" --remote-debugging-port=9222

# macOS
/Applications/TradingView.app/Contents/MacOS/TradingView --remote-debugging-port=9222

# Linux
tradingview --remote-debugging-port=9222
```

### Проверка подключения

```bash
tv status
tv doctor
```

---

## Настройка MCP

Добавьте `tvmcp` в конфигурацию вашего MCP-клиента.

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

### Cursor (`~/.cursor/mcp.json` или `%APPDATA%\Cursor\User\mcp.json`)

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

Используйте тот же формат `mcpServers` — он стандартный для MCP-протокола.  
Пути к config-файлам каждого клиента: `scripts/configure-mcp.sh --list`

### Автоматическая настройка

```bash
# Linux / macOS
bash scripts/configure-mcp.sh --client claude
bash scripts/configure-mcp.sh --client cursor

# Windows
.\scripts\configure-mcp.ps1 -Client claude
.\scripts\configure-mcp.ps1 -Client cursor
```

### Проверка после настройки

```bash
tv status        # CDP подключение
tv discover      # доступные API пути
```

`tv discover` также возвращает `compatibility_probes` для undocumented TradingView internals. После обновлений TradingView Desktop проверьте эти probes перед использованием study model values, strategy reports или equity extraction. Strategy equity остаётся `coverage: loaded_chart_bars`; derived equity conditional и не native Strategy Tester equity.

---

## Поддерживаемые AI-провайдеры

`tvmcp` — стандартный MCP-сервер поверх **stdio** (JSON-RPC 2.0). Он **не привязан** к Claude Code или Anthropic.

| Клиент | Поддержка MCP |
| --- | --- |
| **Claude Code** (Anthropic CLI) | Да — основная платформа разработки |
| **Cursor** | Да — через MCP settings |
| **Cline** (VS Code extension) | Да |
| **Continue** (VS Code extension) | Да |
| **Windsurf** | Да |
| **OpenAI Codex** | Только если реализует MCP stdio |
| Любой кастомный клиент | Да, если поддерживает JSON-RPC 2.0 / stdio MCP |

Сервер не делает сетевых вызовов к AI-провайдерам — он только предоставляет инструменты.
