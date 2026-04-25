# Навыки и агенты

> [← Назад к документации](README.md)

---

## Навыки (Skills)

Директория `skills/` содержит готовые workflow-сценарии для AI-ассистентов.

| Навык | Описание |
| --- | --- |
| `chart-analysis` | Технический анализ: символ, индикаторы, разметка, скриншот |
| `multi-symbol-scan` | Скан нескольких символов, batch-сравнение |
| `pine-develop` | Разработка Pine Script: написание → компиляция → исправление ошибок |
| `replay-practice` | Ручная торговля в режиме Replay |
| `strategy-report` | Отчёт о производительности стратегии |

Чтобы использовать навык, откройте `skills/<name>/SKILL.md` в контексте вашего AI-клиента.

---

## Агенты

Директория `agents/` содержит **нативные форматы** для каждого AI-клиента — без дополнительной конвертации.

### `performance-analyst`

Анализирует результаты бэктеста: метрики, сделки, кривую доходности → структурированный отчёт.

Автоматически вызывает: `data_get_strategy_results`, `data_get_trades`, `data_get_equity`,
`chart_get_state`, `capture_screenshot`.

| Клиент | Файл | Установка |
| --- | --- | --- |
| **Claude Code** | `agents/performance-analyst.md` | `claude --agent agents/performance-analyst.md` |
| **Cursor** | `agents/cursor/performance-analyst.mdc` | скопировать в `.cursor/rules/` |
| **Cline** | `agents/cline/performance-analyst.md` | скопировать в `.clinerules/` |
| **Windsurf** | `agents/windsurf/performance-analyst.md` | добавить в `.windsurfrules` |
| **Continue** | `agents/continue/performance-analyst.prompt` | скопировать в `.continue/prompts/` |
| **OpenAI Codex CLI** | `agents/codex/performance-analyst.md` | скопировать в `AGENTS.md` или `--instructions` |
| **Gemini CLI** | `agents/gemini/performance-analyst.md` | скопировать в `GEMINI.md` или `--system` |

Подробные инструкции установки для каждого клиента: [agents/README.md](../../agents/README.md)

Универсальный системный промпт (источник истины): [prompts/performance-analyst.md](../../prompts/performance-analyst.md)
