package data

import (
	"encoding/json"
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
		"data_get_strategy_results",
		"data_get_trades",
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
