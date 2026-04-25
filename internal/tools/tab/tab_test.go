package tab

import (
	"testing"

	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
)

func TestRegisterTabToolNames(t *testing.T) {
	reg := mcp.NewRegistry()
	RegisterTools(reg)
	tools := reg.List()
	names := make(map[string]bool, len(tools))
	for _, td := range tools {
		names[td.Name] = true
	}
	for _, want := range []string{"tab_list", "tab_new", "tab_close", "tab_switch"} {
		if !names[want] {
			t.Errorf("tool %q not registered", want)
		}
	}
}

func TestSwitchTabEmptyID(t *testing.T) {
	result, err := SwitchTab("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["success"] == true {
		t.Error("expected success=false for empty tab_id")
	}
}
