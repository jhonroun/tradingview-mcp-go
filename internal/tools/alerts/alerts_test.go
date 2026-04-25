package alerts

import (
	"testing"

	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
)

func TestRegisterAlertToolNames(t *testing.T) {
	reg := mcp.NewRegistry()
	RegisterTools(reg)

	want := []string{
		"alert_create",
		"alert_list",
		"alert_delete",
		"watchlist_get",
		"watchlist_add",
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

func TestDeleteAlertsIndividualError(t *testing.T) {
	_, err := DeleteAlerts(false)
	if err == nil {
		t.Error("expected error when delete_all=false")
	}
}
