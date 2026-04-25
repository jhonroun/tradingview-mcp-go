# ORCHESTRATION_AND_LIMITS.md — оркестрация портирования в лимитах Claude Pro

## Цель

Организовать перенос `tradingview-mcp` на Go так, чтобы не сжечь лимиты Claude Pro и не потерять качество.

## Основной принцип

Claude не должен каждый раз заново читать весь репозиторий. Контекст нужно сжимать в рабочие артефакты:

- `PLAN.md` — стратегический план;
- `TODO.md` — текущие задачи;
- `CHANGELOG.md` — что реально изменено;
- `COMPATIBILITY_MATRIX.md` — соответствие Node.js → Go;
- `PORTING_NOTES.md` — короткие факты, найденные при чтении исходников;
- `TEST_REPORT.md` — результаты запусков.

## Режим работы в лимитах

### Сессия 1 — инвентаризация

Задача: не писать код, а снять карту исходника.

Результат:

- список файлов;
- список MCP tools;
- список CLI-команд;
- карта CDP-вызовов;
- карта skills/utilities/scripts;
- риски WindowsApps;
- заполненный `COMPATIBILITY_MATRIX.md`.

### Сессия 2 — Go skeleton

Задача: создать компилируемый Go-каркас без бизнес-логики.

Результат:

- `go.mod`;
- `cmd/tvmcp`;
- `cmd/tv`;
- `internal/mcp`;
- `internal/cdp`;
- `internal/discovery`;
- `internal/launcher`;
- базовые тесты.

### Сессия 3 — CDP + launch

Задача: подключиться к TradingView Desktop и пройти health-check.

Результат:

- CDP client;
- Windows/Mac/Linux launcher;
- WindowsApps discovery;
- `tv_health_check`;
- `tv doctor windows`.

### Сессии 4+ — перенос группами

Переносить строго по группам:

1. health/session/version;
2. chart read-only;
3. chart navigation;
4. symbol/timeframe/layout;
5. screenshots;
6. Pine workflow;
7. drawings/alerts/replay;
8. skills/utilities;
9. CLI parity;
10. test parity.

## Правило одного окна

Одна сессия Claude = одна группа задач. Запрещено смешивать:

- MCP registry;
- CDP runtime JS;
- CLI formatting;
- discovery/launcher;
- tests.

## Как экономить контекст

В начале каждой новой сессии давать Claude только:

1. `CLAUDE.md`;
2. актуальный фрагмент `TODO.md`;
3. нужный фрагмент `PORTING_GUIDE.md`;
4. 1–3 исходных Node.js-файла;
5. 1–3 целевых Go-файла;
6. текущий compile/test output.

Не давать весь репозиторий без необходимости.

## Делегирование локальной модели

Локальную модель/Qwen/Gemma можно использовать для чернового переноса однотипного кода:

- структуры request/response;
- CLI switch/case;
- таблицы tool registry;
- документация;
- простые тесты.

Claude должен оставаться ревьюером и архитектором:

- проверка контрактов;
- исправление edge cases;
- CDP/Windows/debugging;
- финальное приведение к Go idioms.

## Запрещено

- Поручать локальной модели менять архитектуру.
- Переносить сразу все tools одним огромным патчем.
- Удалять Node.js-файлы до полной parity-проверки.
- Закрывать TODO без теста или ручной проверки.

## Минимальный промпт для каждой сессии

Использовать `PROMPTS.md`.
