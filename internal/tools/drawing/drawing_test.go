package drawing

import (
	"math"
	"testing"

	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
)

func TestRegisterDrawingToolNames(t *testing.T) {
	reg := mcp.NewRegistry()
	RegisterTools(reg)

	want := []string{
		"draw_shape",
		"draw_list",
		"draw_get_properties",
		"draw_remove_one",
		"draw_clear",
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

func TestRequireFinite(t *testing.T) {
	if err := requireFinite(1700000000.0, "ts"); err != nil {
		t.Errorf("expected nil for valid float, got %v", err)
	}
	if err := requireFinite(math.NaN(), "ts"); err == nil {
		t.Error("expected error for NaN")
	}
	if err := requireFinite(math.Inf(1), "ts"); err == nil {
		t.Error("expected error for +Inf")
	}
	if err := requireFinite(math.Inf(-1), "ts"); err == nil {
		t.Error("expected error for -Inf")
	}
}

func TestFmtNum(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{1700000000.0, "1700000000"},
		{25000.5, "25000.5"},
		{0.0, "0"},
		{-3.14, "-3.14"},
	}
	for _, c := range cases {
		if got := fmtNum(c.in); got != c.want {
			t.Errorf("fmtNum(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestDrawShapeValidation(t *testing.T) {
	_, err := DrawShape(DrawShapeArgs{
		Shape: "horizontal_line",
		Point: DrawPoint{Time: math.NaN(), Price: 25000},
	})
	if err == nil {
		t.Error("expected error for NaN time")
	}
}
