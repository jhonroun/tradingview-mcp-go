package data

import (
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
)

// TestRegisteredToolNames verifies the exact tool names expected by the
// Node.js compatibility matrix (no renames allowed).
func TestRegisteredToolNames(t *testing.T) {
	reg := mcp.NewRegistry()
	RegisterTools(reg)

	want := []string{
		"data_get_ohlcv",
		"quote_get",
		"data_get_study_values",
		"data_get_pine_lines",
		"data_get_pine_labels",
		"data_get_pine_tables",
		"data_get_pine_boxes",
		"data_get_indicator",
		"data_get_indicator_history",
		"data_get_strategy_results",
		"data_get_trades",
		"data_get_orders",
		"data_get_equity",
		"depth_get",
	}
	tools := reg.List()
	got := make(map[string]bool, len(tools))
	for _, t := range tools {
		got[t.Name] = true
	}
	for _, name := range want {
		if !got[name] {
			t.Errorf("missing tool: %s", name)
		}
	}
	if len(tools) != len(want) {
		t.Errorf("registered %d tools, want %d", len(tools), len(want))
	}
}

func TestRound2(t *testing.T) {
	cases := []struct{ in, want float64 }{
		{1.2345, 1.23},
		{0, 0},
		{100.0, 100.0},
		{-3.456, -3.46},
	}
	for _, c := range cases {
		if got := round2(c.in); got != c.want {
			t.Errorf("round2(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseDisplayNumberLocales(t *testing.T) {
	cases := []struct {
		in   string
		want float64
	}{
		{"31,51", 31.51},
		{"14,63 K", 14630},
		{"1.2K", 1200},
		{"1,2 M", 1200000},
		{"−3,45", -3.45},
		{"1,234.56", 1234.56},
		{"1.234,56", 1234.56},
		{"1,234", 1234},
		{"17,27 K", 17270},
		{"4.95%", 4.95},
	}
	for _, tc := range cases {
		got, ok := ParseDisplayNumber(tc.in)
		if !ok {
			t.Fatalf("ParseDisplayNumber(%q) unavailable, want %v", tc.in, tc.want)
		}
		if math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("ParseDisplayNumber(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestParseDisplayNumberUnavailable(t *testing.T) {
	for _, in := range []string{"", " ", "—", "na", "N/A", "∅", "--"} {
		if got, ok := ParseDisplayNumber(in); ok {
			t.Errorf("ParseDisplayNumber(%q) = %v, want unavailable", in, got)
		}
	}
}

func TestNormalizeRawStudyPlotsMarksDisplaySource(t *testing.T) {
	plots := normalizeRawStudyPlots([]rawStudyPlot{
		{Name: "RSI", DisplayValue: "31,51"},
		{Name: "Missing", DisplayValue: "—"},
	})
	if len(plots) != 2 {
		t.Fatalf("len(plots) = %d, want 2", len(plots))
	}
	if plots[0].Current == nil || *plots[0].Current != 31.51 {
		t.Fatalf("first plot current = %v, want 31.51", plots[0].Current)
	}
	if plots[0].Values[0] == 3151 {
		t.Fatal("localized decimal comma was parsed as thousands")
	}
	if plots[0].Source != SourceTradingViewUIDataWindow {
		t.Errorf("source = %q, want %q", plots[0].Source, SourceTradingViewUIDataWindow)
	}
	if plots[0].Reliability != ReliabilityDisplayValueLocalizedUIString {
		t.Errorf("reliability = %q, want %q", plots[0].Reliability, ReliabilityDisplayValueLocalizedUIString)
	}
	if plots[0].ReliableForTradingLogic {
		t.Error("UI display strings must not be marked reliable for trading logic")
	}
	if plots[1].Current != nil || len(plots[1].Values) != 0 {
		t.Errorf("unavailable plot = current %v values %v, want null/empty", plots[1].Current, plots[1].Values)
	}
}

func TestNormalizeDepthLevelsUsesDisplayParser(t *testing.T) {
	levels := normalizeDepthLevels([]rawDepthLevel{
		{PriceDisplayValue: "2,51", SizeDisplayValue: "14,63 K"},
		{PriceDisplayValue: "—", SizeDisplayValue: "1"},
	})
	if len(levels) != 1 {
		t.Fatalf("len(levels) = %d, want 1", len(levels))
	}
	if levels[0].Price != 2.51 {
		t.Errorf("price = %v, want 2.51", levels[0].Price)
	}
	if levels[0].Size != 14630 {
		t.Errorf("size = %v, want 14630", levels[0].Size)
	}
	if levels[0].Source != SourceTradingViewUIDOM {
		t.Errorf("source = %q, want %q", levels[0].Source, SourceTradingViewUIDOM)
	}
	if levels[0].ReliableForTradingLogic {
		t.Error("DOM display strings must not be marked reliable for trading logic")
	}
}

func TestBuildStudyModelJSUsesModelPaths(t *testing.T) {
	js := buildStudyModelJS(studyModelQuery{EntityID: "Vvzmzg"}, 10, true, false)
	for _, want := range []string{
		"valueAt",
		"fullRangeIterator",
		"meta.plots",
		"meta.styles",
		SourceTradingViewStudyModel,
		ReliabilityPineRuntimeUnstableInternal,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("study model JS missing %q", want)
		}
	}
	if strings.Contains(js, "dataWindowView") {
		t.Error("study model JS must not use dataWindowView for numeric values")
	}
}

func TestBuildStrategyReportJSUsesBacktestingAPI(t *testing.T) {
	js := buildStrategyReportJS(strategyReportModeSummary, 20)
	for _, want := range []string{
		"await window.TradingViewApi.backtestingStrategyApi()",
		"model.strategySources",
		"model.activeStrategySource",
		"activeStrategyReportData",
		"filledOrders",
		"trades",
		"performance",
		"no_strategy_loaded",
		SourceTradingViewBacktestingAPI,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("strategy report JS missing %q", want)
		}
	}
	if strings.Contains(js, "dataSources()") {
		t.Fatal("strategy report JS must not detect strategies through dataSources()")
	}
}

func TestBuildStrategyReportJSEquityUsesStrategyPlot(t *testing.T) {
	js := buildStrategyReportJS(strategyReportModeEquity, 500)
	for _, want := range []string{
		"fullRangeIterator",
		"Strategy Equity",
		"needs_equity_plot",
		"suggested_pine_line",
		"loaded_chart_bars",
		"timeToMs",
		SourceTradingViewStrategyPlot,
		SourceDerivedFromOHLCVAndTrades,
		ReliabilityPineRuntimeUnstableInternal,
		"plot(strategy.equity",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("equity JS missing %q", want)
		}
	}
	if strings.Contains(js, "native_backtesting_report_equity_series") {
		t.Fatal("equity JS must not expose native report arrays as full equity")
	}
}

func TestClampStrategyLimit(t *testing.T) {
	if got := clampStrategyLimit(0, 20, 20); got != 20 {
		t.Errorf("default limit = %d, want 20", got)
	}
	if got := clampStrategyLimit(999, 20, 50); got != 50 {
		t.Errorf("max limit = %d, want 50", got)
	}
	if got := clampStrategyLimit(7, 20, 50); got != 7 {
		t.Errorf("explicit limit = %d, want 7", got)
	}
}

func TestBuildGraphicsJSContainsFilter(t *testing.T) {
	js := buildGraphicsJS("dwglines", "lines", "MyFilter")
	if js == "" {
		t.Fatal("buildGraphicsJS returned empty string")
	}
	filterJS, _ := json.Marshal("MyFilter")
	if !strings.Contains(js, string(filterJS)) {
		t.Errorf("buildGraphicsJS does not contain JSON-escaped filter %s", filterJS)
	}
}

func TestJoinStr(t *testing.T) {
	if got := joinStr([]string{"a", "b", "c"}, " | "); got != "a | b | c" {
		t.Errorf("joinStr = %q, want %q", got, "a | b | c")
	}
	if got := joinStr(nil, " | "); got != "" {
		t.Errorf("joinStr(nil) = %q, want empty", got)
	}
}

func TestCoalesce(t *testing.T) {
	if coalesce("a", "b") != "a" {
		t.Error("coalesce should return first non-empty string")
	}
	if coalesce("", "b") != "b" {
		t.Error("coalesce should fall back to second string")
	}
}
