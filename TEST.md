# TEST.md — Интеграционное тестирование tradingview-mcp-go

> **Аудитория:** этот файл читает Claude Code с нулевым контекстом.  
> Раздел «Часть 1» выполняет **человек** (один раз). Раздел «Часть 2» — **Claude Code**.

---

## Предварительные условия

- [x] TradingView Desktop запущен с флагом `--remote-debugging-port=9222`  
  _(подтверждается пользователем до начала теста)_
- [x] Порт 9222 отвечает по адресу `localhost:9222`
- [x] На графике открыт символ фьючерса: **MOEX:NG1!** (природный газ, Московская биржа)
- [x] Установлен таймфрейм: **1D** (дневной)
- [x] В репозитории установлен Go ≥ 1.21 (`go version`)
- [x] Рабочая директория — корень репозитория `tradingview-mcp-go`

---

## Часть 1 — Настройка (выполняет человек)

### Шаг 1: Сборка MCP-сервера из исходников

**Windows (PowerShell / bash):**

```bash
go build -o bin/tvmcp.exe ./cmd/tvmcp
go build -o bin/tv.exe    ./cmd/tv
```

**Linux / macOS:**

```bash
go build -o bin/tvmcp ./cmd/tvmcp
go build -o bin/tv    ./cmd/tv
```

Проверка:

```bash
bin/tv status          # ожидаемый ответ: {"connected": true, ...}
```

Если `connected: false` — TradingView не запущен с CDP. Запустите:

```bash
bin/tv launch          # или запустите TradingView вручную с --remote-debugging-port=9222
```

---

### Шаг 2: Регистрация MCP-сервера в Claude Code

Добавьте блок в файл `.claude/settings.json` в корне репозитория (создайте файл если его нет):

**Windows:**

```json
{
  "mcpServers": {
    "tradingview": {
      "command": "bin\\tvmcp.exe",
      "args": [],
      "cwd": "."
    }
  }
}
```

**Linux / macOS:**

```json
{
  "mcpServers": {
    "tradingview": {
      "command": "bin/tvmcp",
      "args": [],
      "cwd": "."
    }
  }
}
```

После сохранения файла перезапустите Claude Code (или выполните `/mcp` для перечитывания конфигурации).  
Убедитесь что `tradingview` присутствует в списке MCP-серверов и статус — `connected`.

---

### Шаг 3: Открытие нужного символа в TradingView

1. В TradingView Desktop откройте или переключитесь на символ **MOEX:NG1!**
2. Установите таймфрейм **1D** (дневной)
3. Желательно добавить базовые индикаторы: RSI, MACD, EMA(20), EMA(50), Volume

---

### Шаг 4: Создание директории результатов

```bash
mkdir -p results
```

---

## Часть 2 — Тестовая сессия (выполняет Claude Code)

> **Инструкция для Claude Code:**  
> Ты проводишь интеграционное тестирование MCP-сервера `tradingview-mcp-go`.  
> Прочитай этот файл полностью, затем выполняй задания по порядку.

---

### Роль и контекст

**Прочитай файл агента:** [`agents/futures-analyst.md`](agents/futures-analyst.md)

Ты выступаешь в роли **futures-analyst** — специалиста по фьючерсам на Московской бирже.  
Объект анализа: **фьючерс на природный газ MOEX:NG1!**, **дневной таймфрейм**.

---

### Обязательные требования к выводу

- Весь вывод — **только на русском языке**
- Каждый результат сохранять в отдельный файл в директории `results/`
- Имя файла = название навыка (например `results/market-brief.md`)
- Каждый файл начинается с заголовка:
  ```
  # [Название навыка] — MOEX:NG1! 1D
  **Дата:** YYYY-MM-DD  
  **Инструменты MCP:** tool1, tool2, ...
  ```
- Если навык пропущен по объективной причине — создать файл с объяснением

---

### Тест 00 — Проверка подключения

**Файл:** `results/00-health-check.md`

Выполни:

1. Вызов `tv_health_check` (или `chart_get_state` если нет health check)
2. Запись версии сервера, статуса CDP, найденного target URL
3. Оценка: подключение работает / не работает

---

### Тест 01 — LLM Context Builder

**Skill:** [`skills/llm-context/SKILL.md`](skills/llm-context/SKILL.md)  
**Файл:** `results/01-llm-context.md`

Выполни навык `llm-context` для MOEX:NG1! 1D:

1. Вызов `chart_context_for_llm` с `top_n: 5`
2. Проверь наличие `symbol`, `timeframe`, `price`, `indicators`, `context_text`
3. Если `indicator_count == 0` — добавь RSI через `chart_manage_indicator`, повтори
4. Сохрани:
   - Полный ответ `context_text`
   - Оценку: все поля присутствуют / что отсутствует
   - Количество индикаторов в ответе

---

### Тест 02 — Рыночный брифинг

**Skill:** [`skills/market-brief/SKILL.md`](skills/market-brief/SKILL.md)  
**Файл:** `results/02-market-brief.md`

Выполни навык `market-brief` для MOEX:NG1! 1D:

1. `market_summary` — ценовые данные, объём, все индикаторы
2. Классификация объёма по `volume_vs_avg`
3. Для каждого индикатора: `indicator_state { "name": "..." }`
4. Опционально: `capture_screenshot { "region": "chart" }`
5. Вывод в стандартном формате из SKILL.md:
   ```
   ## Рыночный брифинг — MOEX:NG1! | 1D | [дата]
   **Ценовое действие** ...
   **Объём** ...
   **Индикаторы** ...
   **Вывод** ...
   ```

---

### Тест 03 — Скан индикаторов

**Skill:** [`skills/indicator-scan/SKILL.md`](skills/indicator-scan/SKILL.md)  
**Файл:** `results/03-indicator-scan.md`

Выполни навык `indicator-scan`:

1. `chart_get_state` — список активных индикаторов
2. `indicator_state` для каждого (RSI, MACD, EMA и другие с графика)
3. Таблица сигналов: индикатор / значение / сигнал / направление
4. Оценка конфлюэнтности: бычий / медвежий / смешанный
5. Особое внимание на `near_zero: true` — потенциальные пересечения

---

### Тест 04 — Анализ непрерывного контракта

**Skill:** [`skills/futures-roll/SKILL.md`](skills/futures-roll/SKILL.md)  
**Файл:** `results/04-futures-roll.md`

Выполни навык `futures-roll` для NG1!:

1. `continuous_contract_context` — определение `base_symbol`, `roll_number`, `exchange`
2. `market_summary` для оценки объёма (индикатор периода ролла)
3. Классификация по `volume_vs_avg`: нормально / приближение ролла / активный ролл
4. Проверка контракта NG2! через `quote_get` (спред контанго/бэквордация)
5. Вывод в формате:
   ```
   ## Фьючерсный контракт: MOEX:NG1!
   **Базовый символ:** NG  
   **Номер ролла:** 1 (ближний месяц)
   **Биржа / Тип:** ...
   **Статус ролла:** ...
   **Спред:** ...
   **Рекомендация:** ...
   ```

---

### Тест 05 — Технический анализ графика

**Skill:** [`skills/chart-analysis/SKILL.md`](skills/chart-analysis/SKILL.md)  
**Файл:** `results/05-chart-analysis.md`

Выполни навык `chart-analysis`:

1. Убедись что символ MOEX:NG1! и таймфрейм 1D установлены
2. `data_get_ohlcv` с `count: 50` — последние 50 дневных баров
3. `quote_get` — текущая цена, изменение, объём
4. `symbol_info` — метаданные символа (тип, биржа, сессия)
5. `capture_screenshot { "region": "chart" }` — скриншот графика
6. `data_get_pine_lines` — уровни поддержки/сопротивления из индикаторов (если есть)
7. Анализ:
   - Текущая цена и диапазон за 50 баров
   - Ключевые уровни поддержки и сопротивления
   - Показания индикаторов (RSI: перекупленность/перепроданность)
   - Общий бычий/медвежий/нейтральный уклон с обоснованием

---

### Тест 06 — Мультисимвольное сканирование

**Skill:** [`skills/multi-symbol-scan/SKILL.md`](skills/multi-symbol-scan/SKILL.md)  
**Файл:** `results/06-multi-symbol-scan.md`

Выполни навык `multi-symbol-scan` для энергетических фьючерсов:

Символы: `MOEX:NG1!`, `NYMEX:NG1!`, `NYMEX:CL1!`  
Таймфрейм: 1D

Для каждого символа:
1. `chart_set_symbol` → `market_summary` → сохрани ключевые данные
2. Обязательные поля: symbol, last, change_pct, volume_vs_avg

Итоговая таблица сравнения:

| Символ | Цена | Изм.% | Объём/Ср | RSI | Уклон |
|--------|------|-------|----------|-----|-------|
| ... | ... | ... | ... | ... | ... |

---

### Тест 07 — Верификация JSON-контрактов

**Skill:** [`skills/json-contracts/SKILL.md`](skills/json-contracts/SKILL.md)  
**Файл:** `results/07-json-contracts.md`

Выполни навык `json-contracts` — проверь соответствие Phase 5 контрактам:

Верни граф к MOEX:NG1! 1D перед проверкой.

1. **`data_get_study_values`** — проверить наличие `entity_id`, `plot_count`, `plots` у каждого индикатора; `studies` не null
2. **`chart_get_state`** — проверить наличие `exchange`, `ticker`, `pane_count`, `indicators`
3. **`quote_get`** — проверить наличие `bid`, `ask`, `change`, `change_pct` (все числа, не null)
4. **`symbol_info`** — проверить наличие `symbol`, `exchange`, `description`, `type`
5. **`symbol_search { "query": "NG" }`** — проверить что каждый результат имеет все 4 поля
6. **`data_get_indicator`** — вызвать с entity_id первого индикатора из `chart_get_state`; проверить `inputs` как объект, `plots` как массив

Для каждой проверки:
- ✅ или ❌ рядом с полем
- Фактическое значение поля (первые 50 символов)
- Итоговый статус: PASS / FAIL

---

### Тест 08 — Обработка ошибок

**Skill:** [`skills/error-handling/SKILL.md`](skills/error-handling/SKILL.md)  
**Файл:** `results/08-error-handling.md`

Выполни навык `error-handling` — проверь классификацию ошибок:

1. **Неизвестный инструмент** — вызови любой несуществующий инструмент через MCP (если это не разрешено прямо, опиши ожидаемое поведение согласно SKILL.md)
2. **Отсутствующий обязательный параметр** — вызови `data_get_indicator` без `entity_id`; получи `"entity_id is required"`
3. **Частичное имя индикатора** — `indicator_state { "name": "НЕСУЩЕСТВУЮЩИЙ_ИНДИКАТОР_XYZ" }`; получи `success: false`

Для каждого случая:
- Фактическое сообщение об ошибке
- Классификация по таблице из SKILL.md (повторяемая / постоянная)
- Соответствует ли поведение контракту: ДА / НЕТ

---

### Тест 09 — Отчёт по стратегии (условный)

**Skill:** [`skills/strategy-report/SKILL.md`](skills/strategy-report/SKILL.md)  
**Файл:** `results/09-strategy-report.md`

```
УСЛОВИЕ ВЫПОЛНЕНИЯ: на графике MOEX:NG1! должна быть загружена Pine Script стратегия
(не просто индикатор). Проверь через `data_get_strategy_results`.
```

Если стратегия есть:
1. `data_get_strategy_results` — метрики
2. `data_get_trades` с `max_trades: 10`
3. `capture_screenshot { "region": "strategy_tester" }`
4. Полный отчёт по шаблону из SKILL.md

Если стратегии нет:
- Создай файл с текстом: «Стратегия не загружена. Тест пропущен. Для выполнения добавьте любую Pine Script стратегию на график MOEX:NG1!»

---

### Тест 10 — Pine Script (справочный)

**Skill:** [`skills/pine-develop/SKILL.md`](skills/pine-develop/SKILL.md)  
**Файл:** `results/10-pine-develop.md`

```
УСЛОВИЕ: этот тест требует интерактивной работы с редактором Pine Script.
Создай файл с описанием того, что навык умеет делать применительно к NG1!,
и что нужно пользователю чтобы запустить его полноценно.
```

Создай файл с разделами:
- Что делает навык `pine-develop`
- Применение для MOEX:NG1!: примеры скриптов которые могут быть полезны (RSI с адаптивными уровнями, трекер ролла контракта)
- Команды MCP которые использует навык
- Как запустить интерактивную часть вручную

---

### Тест 11 — Replay (справочный)

**Skill:** [`skills/replay-practice/SKILL.md`](skills/replay-practice/SKILL.md)  
**Файл:** `results/11-replay-practice.md`

Аналогично тесту 10 — навык требует активного режима Replay в TradingView.

Создай справочный файл:
- Что делает навык `replay-practice`
- Как применить к MOEX:NG1! (выбор дат для воспроизведения)
- Команды MCP которые использует навык
- Как активировать Replay в TradingView вручную

---

## Финальный шаг — Сводный отчёт

**Файл:** `results/SUMMARY.md`

После выполнения всех тестов создай сводный файл:

```markdown
# Итоги тестирования tradingview-mcp-go
**Дата:** YYYY-MM-DD  
**Символ:** MOEX:NG1!  
**Таймфрейм:** 1D  
**MCP-сервер:** запущен из ./bin/tvmcp[.exe]

## Статус тестов

| № | Навык | Статус | Файл | Примечание |
|---|-------|--------|------|------------|
| 00 | Health Check | ✅/❌ | 00-health-check.md | |
| 01 | LLM Context | ✅/❌ | 01-llm-context.md | |
| 02 | Market Brief | ✅/❌ | 02-market-brief.md | |
| 03 | Indicator Scan | ✅/❌ | 03-indicator-scan.md | |
| 04 | Futures Roll | ✅/❌ | 04-futures-roll.md | |
| 05 | Chart Analysis | ✅/❌ | 05-chart-analysis.md | |
| 06 | Multi-Symbol Scan | ✅/❌ | 06-multi-symbol-scan.md | |
| 07 | JSON Contracts | ✅/❌ | 07-json-contracts.md | |
| 08 | Error Handling | ✅/❌ | 08-error-handling.md | |
| 09 | Strategy Report | ✅/⏭️ | 09-strategy-report.md | пропущен/выполнен |
| 10 | Pine Develop | ⏭️ | 10-pine-develop.md | справочный |
| 11 | Replay Practice | ⏭️ | 11-replay-practice.md | справочный |

## Текущая оценка рынка NG1!

[2-3 предложения: общий вывод об анализе природного газа на MOEX на дневном таймфрейме,
опираясь на данные из тестов 01–06]

## Выявленные проблемы

[Если были ошибки или несоответствия контрактам — перечислить здесь]

## Заключение

[Работает ли MCP-сервер корректно? Какие навыки требуют доработки?]
```

---

## Справочник: символы природного газа

| Символ | Биржа | Описание |
| --- | --- | --- |
| `MOEX:NG1!` | Московская биржа | Непрерывный контракт NG (ближний месяц) |
| `MOEX:NG2!` | Московская биржа | Второй месяц (для спреда) |
| `NYMEX:NG1!` | NYMEX (CME Group) | Henry Hub NG (мировой бенчмарк) |

При переключении символов используй `chart_set_symbol` — инструмент ожидает загрузки автоматически.

---

## Справочник: команды для диагностики

```bash
# Проверка CDP
bin/tv status

# Диагностика установки (Windows)
bin/tv doctor

# Список всех MCP-инструментов (должно быть 82)
bin/tv --help

# Проверка конкретного инструмента
bin/tv quote MOEX:NG1!
bin/tv chart-state
bin/tv context --top-n 5
bin/tv futures-context
```

---

## Завершение

После создания `results/SUMMARY.md` тестовая сессия завершена.  
Файлы в директории `results/` являются артефактами этого прогона и не должны коммититься в репозиторий.  
Добавьте `results/` в `.gitignore` если нужно.
