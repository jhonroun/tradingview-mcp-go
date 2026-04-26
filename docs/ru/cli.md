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

# HTS — составные команды для LLM (Фаза 4)
tv context [--top-n N]             # состояние графика + цена + top-N индикаторов в одном вызове
tv indicator NAME                  # текущее значение + сигнал для указанного индикатора
tv market                          # полная рыночная сводка (OHLCV, change%, объём vs среднее, индикаторы)
tv futures-context                 # данные непрерывного контракта (NG1!, ES1!, CL2!, …)

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

## `tv doctor` — диагностика Windows

`tv doctor` проверяет локальную систему и возвращает структурированный JSON с рекомендациями. Используйте, когда `tv status` возвращает ошибку и нужно выяснить причину.

```bash
tv doctor
```

### Поля вывода

| Поле | Тип | Описание |
| --- | --- | --- |
| `port.reachable` | bool | `true` если `localhost:9222` отвечает |
| `port.cdp` | bool | `true` если ответ — корректный CDP-список целей |
| `port.owner` | string | Имя процесса, занявшего порт 9222 (например `"chrome.exe"`) |
| `port.error` | string | Причина недоступности в читаемом виде |
| `process.running` | bool | `true` если `TradingView.exe` найден в списке процессов |
| `process.pid` | int | PID запущенного процесса |
| `process.has_cdp_flag` | bool | `true` если в командной строке есть `--remote-debugging-port` |
| `process.cmdline` | string | Полная командная строка процесса |
| `install.found` | bool | `true` если исполняемый файл TradingView найден |
| `install.path` | string | Абсолютный путь к `TradingView.exe` |
| `install.source` | string | Откуда найден (`LOCALAPPDATA`, `Microsoft Store` и т.д.) |
| `install.is_msix` | bool | `true` для установки из Microsoft Store (WindowsApps) |
| `install.local_appdata_dir` | string | Путь к `%LOCALAPPDATA%\TradingView`, если существует |
| `install.appdata_dir` | string | Путь к `%APPDATA%\TradingView`, если существует |
| `launch_cmd` | string | Точная команда для запуска TradingView с CDP-флагом |
| `hints` | string[] | Упорядоченные рекомендации по устранению проблем |

### Пример — TradingView не запущен

```json
{
  "port":    { "reachable": false, "cdp": false, "error": "connection refused — port not listening" },
  "process": { "running": false },
  "install": { "found": true, "path": "C:\\Users\\you\\AppData\\Local\\TradingView\\TradingView.exe", "source": "LOCALAPPDATA" },
  "launch_cmd": "cd \"C:\\Users\\you\\AppData\\Local\\TradingView\" && \"TradingView.exe\" --remote-debugging-port=9222",
  "hints": ["TradingView is not running. Start it: tv launch"]
}
```

### Пример — порт 9222 занят Chrome

```json
{
  "port":    { "reachable": false, "cdp": false, "owner": "chrome.exe", "error": "port in use by \"chrome.exe\" but not CDP" },
  "process": { "running": false },
  "install": { "found": true, "path": "...", "source": "LOCALAPPDATA" },
  "launch_cmd": "...",
  "hints": ["Port 9222 is in use by \"chrome.exe\". Close it or choose a different port."]
}
```

### Пример — TradingView запущен без CDP

```json
{
  "port":    { "reachable": false, "cdp": false, "error": "connection refused — port not listening" },
  "process": { "running": true, "pid": 4812, "has_cdp_flag": false, "cmdline": "TradingView.exe" },
  "install": { "found": true, "path": "...", "source": "LOCALAPPDATA" },
  "launch_cmd": "...",
  "hints": [
    "TradingView.exe is running but --remote-debugging-port is not set. Restart it: tv launch --kill",
    "Or restart manually: cd \"...\" && \"TradingView.exe\" --remote-debugging-port=9222"
  ]
}
```

> Поля `process.*` и `install.*` — только для Windows. На macOS/Linux они пусты; используйте `tv status` и `tv launch`.

---

## HTS-инструменты для LLM (Фаза 4)

Четыре составных инструмента, которые сокращают количество вызовов при работе с LLM:
каждый объединяет несколько базовых MCP-инструментов в один запрос.

### `chart_context_for_llm`

Агрегирует `chart_get_state` + `quote_get` + top-N значений индикаторов в один объект
и формирует строку `context_text` для прямой вставки в промпт LLM.

#### Аргументы

| Поле | Тип | Умолч. | Описание |
| --- | --- | --- | --- |
| `top_n` | integer | 5 | Макс. количество индикаторов в ответе |

#### Поля ответа

| Поле | Тип | Описание |
| --- | --- | --- |
| `symbol` | string | Текущий символ графика |
| `timeframe` | string | Текущий таймфрейм (например `"D"`) |
| `chart_type` | string | Код типа графика |
| `price` | object | `{last, open, high, low, close, volume}` — последний бар |
| `indicators` | array | Top-N объектов с полями `name` и `values` |
| `indicator_count` | int | Количество включённых индикаторов |
| `context_text` | string | `"Symbol: X \| TF: D \| Price: 150 \| RSI(RSI): 65.3 \| …"` |

#### CLI

```bash
tv context              # top 5 индикаторов (по умолчанию)
tv context --top-n 10   # top 10 индикаторов
```

---

### `indicator_state`

Находит индикатор по частичному совпадению имени и классифицирует его текущее значение
в виде направления и сигнала — чтобы LLM не разбирал сырые массивы.

#### Аргументы

| Поле | Тип | Описание |
| --- | --- | --- |
| `name` | string | Частичное имя индикатора (регистр не важен): `"RSI"`, `"MACD"` и т.д. |

#### Поля ответа

| Поле | Тип | Описание |
| --- | --- | --- |
| `matched_name` | string | Полное имя найденного индикатора |
| `values` | object | Все значения из окна данных на текущем баре |
| `primary_value` | number | Первое числовое значение, округлённое до 2 знаков |
| `primary_key` | string | Имя поля с первичным значением |
| `direction` | string | `above_zero` / `below_zero` / `at_zero` |
| `signal` | string | `bullish` / `bearish` / `neutral` / `overbought` / `oversold` |
| `near_zero` | bool | `true` если `\|value\| < 0.5` (индикатор вблизи пересечения нуля) |

Правила классификации:

- RSI / Relative Strength Index / Stochastic: overbought ≥ 70, oversold ≤ 30, иначе neutral
- CCI: overbought ≥ 100, oversold ≤ −100, иначе neutral
- Все остальные: положительное = bullish, отрицательное = bearish, ноль = neutral

#### CLI

```bash
tv indicator RSI
tv indicator "MACD"
tv indicator "Bollinger"
```

---

### `market_summary`

Полный рыночный контекст за один вызов: символ, таймфрейм, OHLCV последнего бара,
изменение цены в %, объём относительно среднего за 20 баров, все активные индикаторы.

#### Поля ответа

| Поле | Тип | Описание |
| --- | --- | --- |
| `symbol` | string | Текущий символ |
| `timeframe` | string | Текущий таймфрейм |
| `chart_type` | string | Код типа графика |
| `last_bar` | object | `{time, open, high, low, close, volume}` |
| `change` | number | close − close предыдущего бара (округлено до 2 зн.) |
| `change_pct` | string | Изменение в процентах, например `"1.35%"` |
| `volume_vs_avg` | number | Объём последнего бара ÷ среднее за 20 предыдущих (2 зн.) |
| `indicators` | array | Все активные индикаторы с `name` и `values` |

#### CLI

```bash
tv market
```

---

### `continuous_contract_context`

Определяет, является ли текущий символ непрерывным фьючерсным контрактом
(`NG1!`, `ES1!`, `CL2!` и т.д.), разбирает базовый символ и номер ролла,
дополняет ответ описанием и биржей из `symbolExt()` TradingView.

#### Поля ответа

| Поле | Тип | Описание |
| --- | --- | --- |
| `symbol` | string | Полный символ с префиксом биржи |
| `is_continuous` | bool | `true` если символ содержит `!` |
| `base_symbol` | string | Корневой символ (например `"NG"` из `"NG1!"`) |
| `roll_number` | int | Номер ролла (1 = ближний, 2 = следующий, …) |
| `description` | string | Человекочитаемое название из TradingView |
| `exchange` | string | Биржа |
| `type` | string | Тип инструмента (например `"futures"`) |
| `currency_code` | string | Валюта расчётов |
| `root_description` | string | Описание корня фьючерса (если доступно) |
| `note` | string | Напоминание: дата экспирации через JS API недоступна |

#### CLI

```bash
tv futures-context
```

---

## JSON-контракты и обработка ошибок (Фаза 5)

Фаза 5 фиксирует схемы ответов для шести инструментов, потребляемых HTS-слоем,
и вводит единую классификацию ошибок.

### Стабильные контракты ответов

#### `data_get_study_values`

```json
{
  "success": true,
  "study_count": 2,
  "studies": [
    {
      "name": "RSI",
      "entity_id": "Study_RSI_0",
      "plot_count": 1,
      "plots": [{ "name": "RSI", "current": 55.3, "values": [55.3] }]
    }
  ]
}
```

- `studies` всегда `[]`, никогда `null` — безопасно итерировать
- `entity_id` — внутренний ID источника данных TradingView (использовать с `data_get_indicator`)
- `plots[0].current === plots[0].values[0]` — алиас текущего значения бара

#### `chart_get_state`

```json
{
  "success": true,
  "symbol": "BINANCE:BTCUSDT",
  "exchange": "BINANCE",
  "ticker": "BTCUSDT",
  "timeframe": "60",
  "type": "1",
  "indicators": [{ "id": "Study_RSI_0", "name": "RSI" }],
  "pane_count": 2
}
```

- `exchange` и `ticker` извлекаются из `symbol` — всегда строки (пустые, если в символе нет `:`)
- `indicators` — каноническое имя поля; `studies` сохранён как алиас для обратной совместимости
- `pane_count` — количество видимых панелей графика

#### `quote_get`

```json
{
  "success": true,
  "symbol": "BINANCE:BTCUSDT",
  "last": 67400.0,
  "open": 66800.0, "high": 67900.0, "low": 66500.0, "close": 67400.0,
  "volume": 12345.67,
  "bid": 0, "ask": 0,
  "change": 600.0, "change_pct": 0.9
}
```

- `bid`, `ask`, `change`, `change_pct` — **всегда присутствуют**; `0` как сентинел для недоступных значений
- `change` = close − close предыдущего бара; `change_pct` — в процентах (не дробь)

#### `symbol_info`

- `symbol`, `exchange`, `description`, `type` — всегда присутствуют (пустая строка, если TradingView не вернул)

#### `symbol_search`

- Каждый результат всегда содержит `symbol`, `exchange`, `description`, `type`

#### `data_get_indicator`

```json
{
  "success": true,
  "entity_id": "Study_RSI_0",
  "name": "Relative Strength Index",
  "inputs": { "length": 14, "source": "close" },
  "plots": [{ "name": "RSI", "current": 55.3, "values": [55.3] }]
}
```

- `inputs` — всегда **объект** ключ→значение (не массив); объёмные строки усекаются
- `plots` — всегда массив (пустой, если у индикатора нет видимых выходов)
- `name` — всегда строка (пустая, если metaInfo недоступен)

---

### Классификация ошибок

Каждый инструмент возвращает либо `{ "success": true, …поля… }`,
либо `{ "success": false, "error": "…" }`.

#### Повторяемые ошибки (транзиентные — ждать и повторить)

| `error` содержит | Причина | Действие |
| --- | --- | --- |
| `"CDP"` или `"connect"` | Chrome DevTools Protocol недоступен | Запустить `tv launch` или TradingView вручную |
| `"no TradingView"` | Вкладка с графиком не найдена | Открыть график в TradingView |
| `"timeout"` | Истёк таймаут выполнения JS | Повторить через 5 с |
| `"websocket"` / `"WebSocket"` | WebSocket-соединение разорвано | Автоматически переподключается; повторить через 3 с |

#### Постоянные ошибки (повторять не нужно)

| `error` содержит | Причина | Действие |
| --- | --- | --- |
| `"unknown tool"` | Имя инструмента не существует | Проверить имя |
| `"unmarshal"` или `"invalid"` | Неверный тип аргумента | Исправить аргументы |
| `"is required"` | Отсутствует обязательное поле | Добавить поле |

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
