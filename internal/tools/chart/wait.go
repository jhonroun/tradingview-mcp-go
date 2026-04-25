package chart

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/jhonroun/tradingview-mcp-go/internal/cdp"
	tv "github.com/jhonroun/tradingview-mcp-go/internal/tradingview"
)

// waitForChartReady polls until the chart finishes loading and shows the
// expected symbol (or any symbol when symbol == ""). Returns false on timeout.
func waitForChartReady(ctx context.Context, client *cdp.Client, symbol string) bool {
	const expr = `(function() {
		var loading = !!document.querySelector('.chart-container.active .chart-status-loading');
		var sym = '';
		var bars = -1;
		try { sym = ` + tv.ChartAPI + `.symbol() || ''; } catch(e) {}
		try { bars = ` + tv.BarsPath + `.size(); } catch(e) {}
		return { loading: loading, symbol: sym, bars: bars };
	})()`

	prev := -1
	stable := 0
	deadline := time.Now().Add(10 * time.Second)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return false
		default:
		}
		time.Sleep(200 * time.Millisecond)

		raw, err := client.Evaluate(ctx, expr, false)
		if err != nil {
			continue
		}
		var st struct {
			Loading bool   `json:"loading"`
			Symbol  string `json:"symbol"`
			Bars    int    `json:"bars"`
		}
		if json.Unmarshal(raw, &st) != nil {
			continue
		}
		if st.Loading {
			prev = -1
			stable = 0
			continue
		}
		if symbol != "" && !strings.Contains(strings.ToLower(st.Symbol), strings.ToLower(symbol)) {
			prev = -1
			stable = 0
			continue
		}
		if st.Bars > 0 && st.Bars == prev {
			stable++
		} else {
			stable = 0
		}
		prev = st.Bars
		if stable >= 2 {
			return true
		}
	}
	return false
}
