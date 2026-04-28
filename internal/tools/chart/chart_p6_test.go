package chart

import (
	"strings"
	"testing"

	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
)

func TestChartTypeMap(t *testing.T) {
	cases := map[string]int{
		"bars": 0, "candles": 1, "line": 2, "area": 3,
		"renko": 4, "kagi": 5, "pointandfigure": 6,
		"linebreak": 7, "heikinashi": 8, "hollowcandles": 9,
	}
	for k, want := range cases {
		got, ok := chartTypeMap[k]
		if !ok {
			t.Errorf("chartTypeMap missing key %q", k)
			continue
		}
		if got != want {
			t.Errorf("chartTypeMap[%q] = %d, want %d", k, got, want)
		}
	}
}

func TestSetTypeUnknown(t *testing.T) {
	_, err := SetType("bogus")
	if err == nil {
		t.Fatal("expected error for unknown chart type")
	}
	if !strings.Contains(err.Error(), "bogus") {
		t.Errorf("error should mention the bad type, got: %v", err)
	}
}

func TestSetTypeNormalization(t *testing.T) {
	// SetType("HeikinAshi") should resolve to chartTypeMap["heikinashi"]=8
	// We can't call CDP in unit tests, but we can test the key normalization logic.
	key := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll("HeikinAshi", " ", ""), "_", ""))
	if key != "heikinashi" {
		t.Errorf("normalization of HeikinAshi = %q, want heikinashi", key)
	}
	key2 := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll("Point And Figure", " ", ""), "_", ""))
	if key2 != "pointandfigure" {
		t.Errorf("normalization of 'Point And Figure' = %q, want pointandfigure", key2)
	}
}

func TestRegisterToolsP6Names(t *testing.T) {
	reg := mcp.NewRegistry()
	RegisterTools(reg)

	want := []string{
		"chart_get_state",
		"chart_get_visible_range",
		"chart_set_symbol",
		"chart_set_timeframe",
		"chart_set_type",
		"chart_manage_indicator",
		"chart_set_visible_range",
		"chart_scroll_to_date",
		"symbol_info",
		"symbol_search",
	}
	got := make(map[string]bool)
	for _, tool := range reg.List() {
		got[tool.Name] = true
	}
	for _, name := range want {
		if !got[name] {
			t.Errorf("missing tool: %s", name)
		}
	}
	if len(reg.List()) != len(want) {
		t.Errorf("registered %d tools, want %d", len(reg.List()), len(want))
	}
}

func TestChartManageIndicatorSchemaIncludesAllowRemoveAny(t *testing.T) {
	reg := mcp.NewRegistry()
	RegisterTools(reg)

	for _, tool := range reg.List() {
		if tool.Name != "chart_manage_indicator" {
			continue
		}
		prop, ok := tool.InputSchema.Properties["allow_remove_any"]
		if !ok {
			t.Fatal("chart_manage_indicator schema missing allow_remove_any")
		}
		if prop.Type != "boolean" {
			t.Fatalf("allow_remove_any type = %q, want boolean", prop.Type)
		}
		return
	}
	t.Fatal("chart_manage_indicator not registered")
}
