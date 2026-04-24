# WINDOWS_TRADINGVIEW_DISCOVERY.md — поиск TradingView Desktop на Windows

## Назначение

Go-порт должен корректно находить и запускать TradingView Desktop на Windows, включая установку через Microsoft Store / WindowsApps.

## Проблема

На Windows TradingView Desktop часто устанавливается не в обычный каталог `%LOCALAPPDATA%\TradingView`, а как packaged app в `C:\Program Files\WindowsApps`, например:

```text
C:\Program Files\WindowsApps\TradingView.Desktop_3.0.0.7652_x64__n534cwy3pjxzj
```

Каталог `WindowsApps` защищён ACL, поэтому прямой обход может завершаться `Access denied`, а путь к exe может меняться при обновлении версии.

## Требования к Go-порту

- Не полагаться только на `%LOCALAPPDATA%\TradingView\TradingView.exe`.
- Поддержать запуск через registered App Execution Alias / shell activation, если прямой exe недоступен.
- Поддержать поиск packaged app через PowerShell/Appx metadata.
- Поддержать пользовательский override через env/config/CLI flag.
- Не падать, если `WindowsApps` недоступен для чтения.
- В диагностике показывать все проверенные способы поиска.

## Приоритет поиска

1. Явно переданный путь:
   - CLI flag: `--tv-path`;
   - env: `TRADINGVIEW_PATH`;
   - config: `tradingview.path`.
2. Уже запущенный TradingView с CDP:
   - проверить `http://127.0.0.1:9222/json/version`;
   - если отвечает — не запускать новый процесс.
3. Стандартная desktop-установка:
   - `%LOCALAPPDATA%\TradingView\TradingView.exe`;
   - `%PROGRAMFILES%\TradingView\TradingView.exe`;
   - `%PROGRAMFILES(X86)%\TradingView\TradingView.exe`.
4. Microsoft Store / WindowsApps:
   - PowerShell `Get-AppxPackage TradingView.Desktop`;
   - чтение `InstallLocation`;
   - поиск executable внутри `InstallLocation`, если доступен.
5. App execution alias / shell:
   - попытка запуска по registered alias, если он есть в `PATH`;
   - fallback через `explorer.exe shell:AppsFolder\<PackageFamilyName>!App`, если нужно только открыть приложение.
6. Ручная ошибка с понятным текстом и примером команды.

## Важное ограничение

`--remote-debugging-port=9222` должен быть передан именно процессу TradingView/Electron. Shell activation может открыть приложение, но не всегда позволяет передать аргументы. Поэтому для packaged app нужно сначала пытаться получить реальный executable path через Appx metadata.

## Диагностический tool

Добавить MCP tool/CLI-команду:

```text
tv doctor windows
```

Она должна вывести:

- ОС и архитектуру;
- найденные процессы TradingView;
- отвечает ли CDP порт 9222;
- список проверенных путей;
- найденный Appx package;
- InstallLocation;
- есть ли права чтения на InstallLocation;
- итоговую команду запуска;
- рекомендации по ручному override.

## Go-модуль

Рекомендуемая структура:

```text
internal/discovery/
  discovery.go
  windows.go
  linux.go
  darwin.go
internal/launcher/
  launcher.go
  windows.go
```

Интерфейс:

```go
type Candidate struct {
    Path       string
    Source     string
    Confidence int
    ArgsOK     bool
    Error      string
}

type Result struct {
    Candidates []Candidate
    Selected   *Candidate
    Diagnostics []string
}
```

## Правило для агентов

Перед изменением CDP-кода не трогать discovery/launcher. Перед изменением discovery/launcher обязательно обновлять `TODO.md`, `CHANGELOG.md` и `COMPATIBILITY_MATRIX.md`.
