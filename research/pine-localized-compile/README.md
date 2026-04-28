# Pine Localized Compile Smoke

Date: 2026-04-27  
OS: Windows  
TradingView UI locale: Russian  
Chart: `RUS:NG1!`, `1D`

## Result

`pine_smart_compile` now recognizes the Russian Add-to-chart button.

Live smoke result:

- `button_clicked`: `Добавить на графикДобавить на график`
- `study_added`: `true`
- `compiled`: `true`
- `error_count`: `0`
- `warning_count`: `1` (`PINE_VERSION_OUTDATED`, expected for the v5 test script)

## Safety

The smoke test used the safe Pine workflow:

1. Captured original Pine source and hash.
2. Set disposable test strategy with `expected-current-sha256`.
3. Ran `pine smart-compile`.
4. Verified test strategy appeared on the chart as `OaWjcc`.
5. Removed only `OaWjcc`.
6. Restored original Pine source from backup.
7. Verified final hash:
   `e675eb546b9f96fb79c1b7cd179d8908e7a77c7fcb4ca798124166042d350765`.

Final chart studies returned to:

- `Vvzmzg` / `Помошник RSI - True ADX`
- `cfyrsD` / `Volume`

## Artifacts

- `before-pine-get.json`
- `before-chart-state.json`
- `set-equity-strategy.json`
- `smart-compile.json`
- `after-smart-compile-chart-state.json`
- `cleanup-remove-test-strategy.json`
- `restore-original.json`
- `after-cleanup-chart-state.json`
- `after-restore-pine-get.json`
