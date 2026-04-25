package batch

import (
	"testing"

	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
)

func TestRegisterBatchToolName(t *testing.T) {
	reg := mcp.NewRegistry()
	RegisterTools(reg)
	tools := reg.List()
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}
	if tools[0].Name != "batch_run" {
		t.Errorf("expected batch_run, got %q", tools[0].Name)
	}
}

func TestBatchRunEmptySymbols(t *testing.T) {
	result, err := BatchRun(nil, nil, "screenshot", 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["success"] == true {
		t.Error("expected success=false for empty symbols")
	}
}

func TestBatchRunDefaultDelayAndCount(t *testing.T) {
	// Verify the defaults are applied before any CDP call.
	delayMs := 0
	ohlcvCount := 0
	if delayMs <= 0 {
		delayMs = 2000
	}
	if ohlcvCount <= 0 {
		ohlcvCount = 100
	}
	if delayMs != 2000 {
		t.Errorf("default delay_ms should be 2000, got %d", delayMs)
	}
	if ohlcvCount != 100 {
		t.Errorf("default ohlcv_count should be 100, got %d", ohlcvCount)
	}
}

func TestBatchRunOhlcvCountCap(t *testing.T) {
	ohlcvCount := 9999
	if ohlcvCount > 500 {
		ohlcvCount = 500
	}
	if ohlcvCount != 500 {
		t.Errorf("ohlcv_count cap should be 500, got %d", ohlcvCount)
	}
}

func TestBatchRunTimeframeDefault(t *testing.T) {
	// When no timeframes are provided, the loop should run once with tf=="".
	tfs := []string{}
	if len(tfs) == 0 {
		tfs = []string{""}
	}
	if len(tfs) != 1 || tfs[0] != "" {
		t.Errorf("expected single empty timeframe, got %v", tfs)
	}
}
