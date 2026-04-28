// Command tvmcp is the MCP stdio server for TradingView.
package main

import (
	"fmt"
	"os"

	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/alerts"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/batch"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/capture"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/chart"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/data"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/drawing"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/health"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/hts"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/indicators"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/pane"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/pine"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/replay"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/tab"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/ui"
)

const serverInstructions = `TradingView MCP — tools for reading and controlling a live TradingView Desktop chart.

TOOL SELECTION GUIDE:

Reading your chart:
- chart_get_state → get symbol, timeframe, all indicator names + entity IDs (call first)
- data_get_study_values → get current numeric values from ALL visible indicators (RSI, MACD, BB, EMA, etc.)
- data_get_indicator_history → get historical study-model values for loaded chart bars
- data_get_strategy_results / data_get_trades / data_get_orders / data_get_equity → read Strategy Tester report when a strategy is loaded
- quote_get → get real-time price snapshot (last, OHLC, volume)
- data_get_ohlcv → get price bars. ALWAYS pass summary=true unless you need individual bars

Reading custom Pine indicator output (line.new/label.new/table.new/box.new drawings):
- data_get_pine_lines → horizontal price levels from custom indicators (deduplicated, sorted)
- data_get_pine_labels → text annotations with prices ("PDH 24550", "Bias Long", etc.)
- data_get_pine_tables → table data as formatted rows (session stats, analytics dashboards)
- data_get_pine_boxes → price zones as {high, low} pairs
- ALWAYS pass study_filter to target a specific indicator by name

Editing Pine source:
- pine_get_source returns source hash/name/type; pine_set_source creates a backup; pine_restore_source verifies SHA256

Screenshots: capture_screenshot → regions: "full", "chart", "strategy_tester"
Launch: tv_launch → auto-detect and start TradingView with CDP on any platform

CONTEXT MANAGEMENT:
- ALWAYS use summary=true on data_get_ohlcv
- ALWAYS use study_filter on pine tools when you know which indicator you want
- Prefer capture_screenshot for visual context over pulling large datasets
- Call chart_get_state ONCE at start, reuse entity IDs`

func main() {
	fmt.Fprintln(os.Stderr, "⚠  tradingview-mcp-go  |  Unofficial tool. Not affiliated with TradingView Inc. or Anthropic.")
	fmt.Fprintln(os.Stderr, "   Ensure your usage complies with TradingView's Terms of Use.")
	fmt.Fprintln(os.Stderr)

	reg := mcp.NewRegistry()

	// P4 — health / launch
	health.RegisterTools(reg)

	// P5 — read-only chart tools; P6 — chart control + symbols (via chart.RegisterTools)
	chart.RegisterTools(reg)
	data.RegisterTools(reg)
	capture.RegisterTools(reg)

	// P6 — indicator control
	indicators.RegisterTools(reg)

	// P7 — Pine Script
	pine.RegisterTools(reg)

	// P8 — Drawing
	drawing.RegisterTools(reg)

	// P9 — Alerts + Watchlist
	alerts.RegisterTools(reg)

	// P10 — Panes + Tabs
	pane.RegisterTools(reg)
	tab.RegisterTools(reg)

	// P11 — Replay
	replay.RegisterTools(reg)

	// P12 — UI automation + Layouts
	ui.RegisterTools(reg)

	// P14 — Batch
	batch.RegisterTools(reg)

	// Phase 4 — HTS-ready composite tools
	hts.RegisterTools(reg)

	srv := mcp.NewServer(reg, serverInstructions)
	if err := srv.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
