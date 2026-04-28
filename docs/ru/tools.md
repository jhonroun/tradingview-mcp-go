# MCP-инструменты (85 текущих Go tools)

> [← Назад к документации](README.md)

Историческая Node.js parity база содержит 78 tools. Текущий Go registry содержит 85 tools: 78 parity tools плюс Go-расширения для истории study, orders, безопасного restore Pine и агрегированных LLM context helpers.

---

### Здоровье и подключение (4)

| Инструмент | Описание |
| --- | --- |
| `tv_health_check` | Проверить CDP подключение |
| `tv_discover` | Показать доступные TradingView API пути |
| `tv_ui_state` | Получить состояние UI |
| `tv_launch` | Запустить TradingView с CDP |

### Состояние графика (2)

| Инструмент | Описание |
| --- | --- |
| `chart_get_state` | Символ, таймфрейм, тип, индикаторы |
| `chart_get_visible_range` | Текущий видимый диапазон дат |

### Управление графиком (6)

| Инструмент | Описание |
| --- | --- |
| `chart_set_symbol` | Сменить символ |
| `chart_set_timeframe` | Сменить таймфрейм |
| `chart_set_type` | Тип графика (свечи, Хейкин-Аши, линия…) |
| `chart_manage_indicator` | Добавить/удалить индикатор; explicit `allow_remove_any` для recovery после лимита |
| `chart_scroll_to_date` | Перейти к дате |
| `chart_set_visible_range` | Установить диапазон видимости |

### Символы (2)

| Инструмент | Описание |
| --- | --- |
| `symbol_info` | Мета-данные символа |
| `symbol_search` | Поиск символов; пустой результат возвращает `status: no_results` |

### Данные (10)

| Инструмент | Описание |
| --- | --- |
| `quote_get` | Текущая котировка OHLCV; unavailable bid/ask помечается `bidAskAvailable:false` |
| `data_get_ohlcv` | Исторические бары |
| `data_get_study_values` | Текущие numeric values видимых studies из TradingView study model, если доступно |
| `data_get_indicator` | Текущие numeric values одного study по entity ID/name |
| `data_get_indicator_history` | История study по loaded bars через `fullRangeIterator()` |
| `data_get_strategy_results` | Метрики стратегии через `TradingViewApi.backtestingStrategyApi()` |
| `data_get_trades` | Сделки стратегии из backtesting report |
| `data_get_orders` | Filled orders стратегии из backtesting report |
| `data_get_equity` | Equity из explicit `Strategy Equity` plot или documented fallback/status |
| `depth_get` | Стакан (Level 2) |

### Pine-графика (4)

| Инструмент | Описание |
| --- | --- |
| `data_get_pine_lines` | Горизонтальные уровни `line.new()` |
| `data_get_pine_labels` | Метки `label.new()` |
| `data_get_pine_tables` | Таблицы `table.new()` |
| `data_get_pine_boxes` | Прямоугольники `box.new()` |

### Скриншот (1)

| Инструмент | Описание |
| --- | --- |
| `capture_screenshot` | Скриншот (full / chart / strategy_tester); `.png` extension нормализуется |

### Pine Script (13)

| Инструмент | Описание |
| --- | --- |
| `pine_get_source` | Прочитать source, hash, script name/type из редактора |
| `pine_set_source` | Создать backup текущего source, затем записать новый source |
| `pine_restore_source` | Восстановить backup и проверить SHA256 |
| `pine_compile` | Compile/add to chart; поддерживает EN/RU Add-to-chart labels |
| `pine_smart_compile` | Compile с diagnostics и проверкой study-added |
| `pine_get_errors` | Structured Monaco errors |
| `pine_get_console` | Вывод Pine-консоли |
| `pine_save` | Сохранить скрипт (Ctrl+S) |
| `pine_new` | Создать новый script (indicator/strategy/library) |
| `pine_open` | Открыть script по имени |
| `pine_list_scripts` | Список сохранённых scripts |
| `pine_analyze` | Offline static analysis |
| `pine_check` | Проверка через pine-facade API |

### Рисование (5)

| Инструмент | Описание |
| --- | --- |
| `draw_shape` | Нарисовать фигуру |
| `draw_list` | Список всех фигур |
| `draw_get_properties` | Свойства фигуры |
| `draw_remove_one` | Удалить фигуру |
| `draw_clear` | Удалить все фигуры |

### Алерты (3)

| Инструмент | Описание |
| --- | --- |
| `alert_create` | Создать ценовой алерт |
| `alert_list` | Список алертов |
| `alert_delete` | Удалить алерты |

### Watchlist (2)

| Инструмент | Описание |
| --- | --- |
| `watchlist_get` | Прочитать watchlist |
| `watchlist_add` | Добавить символ в watchlist |

### Индикаторы (2)

| Инструмент | Описание |
| --- | --- |
| `indicator_set_inputs` | Установить параметры индикатора |
| `indicator_toggle_visibility` | Скрыть/показать индикатор |

### Replay (6)

| Инструмент | Описание |
| --- | --- |
| `replay_start` | Начать replay с даты |
| `replay_step` | Шаг вперёд на один бар |
| `replay_stop` | Остановить replay |
| `replay_status` | Статус (дата, позиция, P&L) |
| `replay_autoplay` | Авто-воспроизведение |
| `replay_trade` | Торговать в replay (buy/sell/close) |

### Панели (4)

| Инструмент | Описание |
| --- | --- |
| `pane_list` | Список pane-ов и layout |
| `pane_set_layout` | Сменить layout |
| `pane_focus` | Сфокусировать pane |
| `pane_set_symbol` | Установить символ для pane |

### Вкладки (4)

| Инструмент | Описание |
| --- | --- |
| `tab_list` | Список вкладок |
| `tab_new` | Открыть новую вкладку |
| `tab_close` | Закрыть вкладку |
| `tab_switch` | Переключиться на вкладку |

### UI-автоматизация (10)

| Инструмент | Описание |
| --- | --- |
| `ui_click` | Клик по элементу |
| `ui_open_panel` | Открыть/закрыть панель |
| `ui_fullscreen` | Полноэкранный режим |
| `ui_keyboard` | Нажатие клавиш |
| `ui_type_text` | Ввод текста |
| `ui_hover` | Наведение курсора |
| `ui_scroll` | Прокрутка |
| `ui_mouse_click` | Клик по координатам |
| `ui_find_element` | Найти элемент на странице |
| `ui_evaluate` | Выполнить JS expression; публичное поведение не менялось |

### Лейауты (2)

| Инструмент | Описание |
| --- | --- |
| `layout_list` | Список сохранённых layouts |
| `layout_switch` | Переключить layout |

### Пакетная обработка (1)

| Инструмент | Описание |
| --- | --- |
| `batch_run` | Обход символов × таймфреймов с действиями |

### LLM/context helpers (4)

| Инструмент | Описание |
| --- | --- |
| `chart_context_for_llm` | Компактный chart state + price + top-N study values |
| `indicator_state` | Signal/direction summary по имени индикатора |
| `market_summary` | OHLCV summary + volume context + active studies |
| `continuous_contract_context` | Метаданные continuous futures и parsing roll-number |

## Политика надёжности данных

Для trading logic нужно проверять `source`, `reliability` и `reliableForTradingLogic`.

- `tradingview_study_model`: numeric Pine runtime values из TradingView study internals; reliable, но unstable internal path.
- `tradingview_backtesting_api`: Strategy Tester report; reliable при `status: ok`, unstable internal path.
- `tradingview_strategy_plot`: explicit Pine `Strategy Equity` plot; reliable только при `coverage: loaded_chart_bars`.
- `tradingview_ui_data_window`: fallback из localized display strings; не reliable для trading logic.
- `derived_from_ohlcv_and_trades`: derived fallback; conditional, `reliableForTradingLogic:false`, если caller отдельно не гарантирует полные OHLCV/trades/settings coverage; не равен native TradingView equity.

## Compatibility probes

`tv_discover` сохраняет legacy объект `paths` и дополнительно возвращает `compatibility_probes`.
Каждый probe не мутирует график и содержит:

- `compatible`: внутренний path/method существует в текущей сборке TradingView.
- `available`: в текущем состоянии графика есть полезные данные.
- `status`: например `ok`, `no_strategy_loaded`, `needs_equity_plot`, `strategy_report_unavailable`, `unavailable`, `error`.
- `stability`: всегда `unstable_internal_path` для undocumented TradingView internals.
- `reliability`: reliability class, который нужно переносить в dependent tool responses.

Запускайте `tv discover` после обновлений TradingView Desktop или когда data tools начинают возвращать unavailable statuses.

## Equity coverage

`data_get_equity` не является полным Strategy Tester equity export. Надёжный путь:

```pine
plot(strategy.equity, "Strategy Equity", display=display.data_window)
```

Если plot есть, tool читает loaded chart bars через strategy source model и возвращает `coverage: loaded_chart_bars`. Это может совпасть с полным диапазоном только если TradingView реально загрузил весь нужный range.

Optional workflow догрузки истории:

1. Использовать `chart_set_visible_range` или `chart_scroll_to_date`, чтобы расширить/сдвинуть диапазон графика.
2. Подождать, пока TradingView догрузит бары.
3. Повторить `data_get_equity` или `data_get_indicator_history`.
4. Сравнить `loaded_bar_count`, `data_points`, `total_data_points`, `coverage`.
5. Сохранять статус loaded-bars coverage, не native full backtest history.

Не тратить implementation time на "full native bar-by-bar Strategy Tester equity", пока TradingView не exposes стабильный report field.
