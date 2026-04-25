package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindViaTradingViewPathEnv(t *testing.T) {
	// Create a temporary fake executable.
	tmp := t.TempDir()
	fake := filepath.Join(tmp, "TradingView.exe")
	if err := os.WriteFile(fake, []byte("fake"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("TRADINGVIEW_PATH", fake)

	result, err := Find()
	if err != nil {
		t.Fatalf("Find() error: %v", err)
	}
	if result.Path != fake {
		t.Errorf("Path = %q, want %q", result.Path, fake)
	}
	if result.Source != "TRADINGVIEW_PATH env" {
		t.Errorf("Source = %q, want TRADINGVIEW_PATH env", result.Source)
	}
}

func TestFindMissingTradingViewPathEnv(t *testing.T) {
	t.Setenv("TRADINGVIEW_PATH", "/nonexistent/path/TradingView.exe")
	// Should fall through to platform search (which will fail in CI),
	// or succeed if TradingView is actually installed.
	// We just verify it doesn't panic and returns coherent result.
	result, err := Find()
	if err == nil && result == nil {
		t.Error("Find() returned nil result and nil error")
	}
	// Either a result or an error — both are acceptable.
}

func TestFileExists(t *testing.T) {
	tmp := t.TempDir()
	existing := filepath.Join(tmp, "exists.txt")
	if err := os.WriteFile(existing, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !fileExists(existing) {
		t.Errorf("fileExists(%q) = false, want true", existing)
	}
	if fileExists(filepath.Join(tmp, "missing.txt")) {
		t.Errorf("fileExists(missing) = true, want false")
	}
}
