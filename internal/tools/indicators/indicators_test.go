package indicators

import (
	"testing"

	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
)

func TestRegisterIndicatorTools(t *testing.T) {
	reg := mcp.NewRegistry()
	RegisterTools(reg)

	want := []string{
		"indicator_set_inputs",
		"indicator_toggle_visibility",
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
