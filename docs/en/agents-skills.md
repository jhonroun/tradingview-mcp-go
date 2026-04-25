# Skills and Agents

> [← Back to docs](README.md)

---

## Skills

The `skills/` directory contains ready-to-use workflow scenarios for AI assistants.

| Skill | Description |
| --- | --- |
| `chart-analysis` | Technical analysis: symbol, indicators, markup, screenshot |
| `multi-symbol-scan` | Multi-symbol scan, batch comparison |
| `pine-develop` | Pine Script development: write → compile → fix errors |
| `replay-practice` | Manual trading practice in Replay mode |
| `strategy-report` | Strategy performance report |

To use a skill, open `skills/<name>/SKILL.md` in your AI client's context.

---

## Agents

The `agents/` directory contains **native-format files** for each AI client — no conversion needed.

### `performance-analyst`

Gathers strategy performance data and produces a structured analysis report.

Automatically calls: `data_get_strategy_results`, `data_get_trades`, `data_get_equity`,
`chart_get_state`, `capture_screenshot`.

| Client | File | Install |
| --- | --- | --- |
| **Claude Code** | `agents/performance-analyst.md` | `claude --agent agents/performance-analyst.md` |
| **Cursor** | `agents/cursor/performance-analyst.mdc` | copy to `.cursor/rules/` |
| **Cline** | `agents/cline/performance-analyst.md` | copy to `.clinerules/` |
| **Windsurf** | `agents/windsurf/performance-analyst.md` | append to `.windsurfrules` |
| **Continue** | `agents/continue/performance-analyst.prompt` | copy to `.continue/prompts/` |
| **OpenAI Codex CLI** | `agents/codex/performance-analyst.md` | copy to `AGENTS.md` or use `--instructions` |
| **Gemini CLI** | `agents/gemini/performance-analyst.md` | copy to `GEMINI.md` or use `--system` |

Full install instructions for each client: [agents/README.md](../../agents/README.md)

Universal system prompt (source of truth): [prompts/performance-analyst.md](../../prompts/performance-analyst.md)
