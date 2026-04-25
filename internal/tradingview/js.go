// Package tradingview holds JS expression constants and string-safety helpers
// for communicating with TradingView Desktop via Chrome DevTools Protocol.
package tradingview

import "encoding/json"

// ChartAPI is the JS path to the active chart widget.
const ChartAPI = `window.TradingViewApi._activeChartWidgetWV.value()`

// BarsPath is the JS path to the main series bars collection.
const BarsPath = ChartAPI + `._chartWidget.model().mainSeries().bars()`

// ChartWidget is the JS path to the chart widget.
const ChartWidget = ChartAPI + `._chartWidget`

// SafeString returns a properly JSON-escaped JS string literal for s.
// Mirrors connection.js safeString() — prevents injection via quotes,
// backticks, template literals, or control characters.
func SafeString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
