# Status

```
tradingview-mcp-go = готов как рабочая база для HTS/MCP-интеграции  
статус: ACCEPTED WITH KNOWN LIMITATIONS
```

Что подтверждено:

```
✅ CDP-подключение к TradingView Desktop работает  
✅ MCP/CLI реально получает данные с графика  
✅ chart\_get\_state работает  
✅ quote\_get работает  
✅ data\_get\_ohlcv работает  
✅ market\_summary работает  
✅ indicator\_state работает  
✅ continuous\_contract\_context работает  
✅ capture\_screenshot работает  
✅ переключение символов работает  
✅ NG1! / NG2! / NYMEX symbols тестировались  
✅ результаты сохранены в файлы
```

Но есть важные ограничения:

```
⚠️ symbol\_search возвращает \[\]  
⚠️ data\_get\_indicator протестирован частично  
⚠️ Pine/strategy/replay — не полноценный live-test, а справочные сценарии  
⚠️ значения индикаторов приходят как canvas/Y-координаты, а не реальные RSI/ADX/EMA  
⚠️ bid/ask = 0 для MOEX-фьючерсов  
⚠️ screenshot filename даёт .png.png
```

Самый важный вывод: **TradingView MCP уже можно использовать как “глаза” системы**, но **нельзя пока использовать его как полноценный источник точных числовых значений кастомных индикаторов**, если они возвращаются в canvas-координатах.

