package pine

import (
	"strings"
	"testing"

	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
)

func TestRegisterPineToolNames(t *testing.T) {
	reg := mcp.NewRegistry()
	RegisterTools(reg)

	want := []string{
		"pine_get_source",
		"pine_set_source",
		"pine_compile",
		"pine_smart_compile",
		"pine_get_errors",
		"pine_get_console",
		"pine_save",
		"pine_new",
		"pine_open",
		"pine_list_scripts",
		"pine_analyze",
		"pine_check",
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

func TestAnalyzeCleanScript(t *testing.T) {
	src := "//@version=6\nindicator(\"Clean\")\nplot(close)"
	result := Analyze(src)
	if result["success"] != true {
		t.Fatal("expected success")
	}
	if result["issue_count"].(int) != 0 {
		t.Errorf("expected 0 issues, got %d", result["issue_count"])
	}
}

func TestAnalyzeArrayOutOfBounds(t *testing.T) {
	src := "//@version=6\nindicator(\"Test\")\narr = array.new_int(3)\nval = array.get(arr, 5)"
	result := Analyze(src)
	diags, _ := result["diagnostics"].([]Diagnostic)
	if len(diags) == 0 {
		t.Fatal("expected out-of-bounds diagnostic")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "out of bounds") && d.Severity == "error" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected out-of-bounds error, got: %v", diags)
	}
}

func TestAnalyzeStrategyWithoutDecl(t *testing.T) {
	src := "//@version=6\nindicator(\"Test\")\nstrategy.entry(\"L\", strategy.long)"
	result := Analyze(src)
	diags, _ := result["diagnostics"].([]Diagnostic)
	if len(diags) == 0 {
		t.Fatal("expected strategy declaration diagnostic")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "strategy()") && d.Severity == "error" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected missing strategy() error, got: %v", diags)
	}
}

func TestAnalyzeOldVersion(t *testing.T) {
	src := "//@version=4\nindicator(\"Old\")\nplot(close)"
	result := Analyze(src)
	diags, _ := result["diagnostics"].([]Diagnostic)
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "v4") && d.Severity == "info" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected old-version info diagnostic, got: %v", diags)
	}
}

func TestAnalyzeStrategyWithDecl(t *testing.T) {
	src := "//@version=6\nstrategy(\"My Strat\", overlay=true)\nstrategy.entry(\"L\", strategy.long)"
	result := Analyze(src)
	diags, _ := result["diagnostics"].([]Diagnostic)
	for _, d := range diags {
		if strings.Contains(d.Message, "strategy()") {
			t.Errorf("should not flag strategy.entry when strategy() exists: %v", d)
		}
	}
}
