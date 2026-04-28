package pine

import (
	"os"
	"path/filepath"
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
		"pine_restore_source",
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

func TestInferScriptMetadataAndHash(t *testing.T) {
	src := "//@version=6\nstrategy(\"MCP Test\", overlay=true)\nplot(close)"
	meta := inferScriptMetadata(src)
	if meta.ScriptType != "strategy" {
		t.Fatalf("ScriptType = %q, want strategy", meta.ScriptType)
	}
	if meta.ScriptName != "MCP Test" {
		t.Fatalf("ScriptName = %q, want MCP Test", meta.ScriptName)
	}
	if meta.PineVersion != "6" {
		t.Fatalf("PineVersion = %q, want 6", meta.PineVersion)
	}
	if sourceSHA256(src) == "" {
		t.Fatal("sourceSHA256 returned empty hash")
	}
}

func TestLoadPineBackupVerifiesSHA256(t *testing.T) {
	dir := t.TempDir()
	source := "//@version=6\nindicator(\"Backup\")\nplot(close)"
	hash := sourceSHA256(source)
	sourcePath := filepath.Join(dir, "backup.pine")
	if err := os.WriteFile(sourcePath, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest := `{"source_sha256":"` + hash + `","source_file":"backup.pine","line_count":3,"char_count":44}`
	manifestPath := filepath.Join(dir, "backup.json")
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}

	loaded, err := loadPineBackup(manifestPath, "")
	if err != nil {
		t.Fatalf("loadPineBackup manifest: %v", err)
	}
	if loaded.SourceSHA256 != hash {
		t.Fatalf("SourceSHA256 = %q, want %q", loaded.SourceSHA256, hash)
	}

	if _, err := loadPineBackup(sourcePath, "wrong"); err == nil {
		t.Fatal("loadPineBackup accepted wrong expected hash")
	}
	if _, err := loadPineBackup(sourcePath, ""); err == nil {
		t.Fatal("loadPineBackup accepted .pine without expected hash")
	}
}

func TestMarkerCounts(t *testing.T) {
	markers := []interface{}{
		map[string]interface{}{"severity_label": "error", "message": "bad"},
		map[string]interface{}{"severity": float64(4), "message": "warn"},
		map[string]interface{}{"severity_label": "info", "message": "note"},
	}
	counts := markerCounts(markers)
	if counts.ErrorCount != 1 {
		t.Fatalf("ErrorCount = %d, want 1", counts.ErrorCount)
	}
	if counts.WarningCount != 1 {
		t.Fatalf("WarningCount = %d, want 1", counts.WarningCount)
	}
}

func TestGetMarkersJSReturnsStructuredDiagnostics(t *testing.T) {
	js := getMarkersJS()
	for _, want := range []string{"severity_label", "end_line", "end_column", "message", "source"} {
		if !strings.Contains(js, want) {
			t.Errorf("getMarkersJS missing %q", want)
		}
	}
}

func TestPineCompileButtonJSRecognizesLocalizedLabels(t *testing.T) {
	for name, js := range map[string]string{
		"compile":       compileButtonJS,
		"smart_compile": smartCompileButtonJS,
	} {
		for _, want := range []string{
			"add to chart",
			"update on chart",
			"save and add to chart",
			"добавить на график",
			"обновить на графике",
			"сохранить и добавить на график",
			"text.slice(0, half) === text.slice(half)",
		} {
			if !strings.Contains(js, want) {
				t.Errorf("%s JS missing %q", name, want)
			}
		}
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
