// Package smoke contains Phase 2 smoke tests that require a live TradingView
// Desktop instance running with CDP enabled on localhost:9222.
//
// All tests call skipIfNoCDP(t) and skip gracefully when TradingView is not
// running. Run manually after starting TradingView with CDP:
//
//	tv launch           # or: set TRADINGVIEW_PATH to a direct-installer build
//	go test ./tests/smoke/... -v -timeout 60s
//
// On Microsoft Store (MSIX) installs the --remote-debugging-port flag is
// suppressed by TradingView's own argument parser. Use the direct-installer
// build from tradingview.com or trigger an in-app relaunch (tv launch --kill).
package smoke

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jhonroun/tradingview-mcp-go/internal/cdp"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/capture"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/chart"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/data"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/health"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/pine"
)

// skipIfNoCDP skips the test if port 9222 is not reachable or has no chart target.
func skipIfNoCDP(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	targets, err := cdp.ListTargets(ctx, "localhost", 9222)
	if err != nil {
		t.Skipf("CDP not available (TradingView not running with --remote-debugging-port=9222): %v", err)
	}
	if _, err := cdp.FindChartTarget(targets); err != nil {
		t.Skipf("TradingView running but no chart target found: %v", err)
	}
}

// ── Phase 2: CDP connect ──────────────────────────────────────────────────────

func TestCDPConnect(t *testing.T) {
	skipIfNoCDP(t)
	result, err := health.HealthCheck()
	if err != nil {
		t.Fatalf("HealthCheck error: %v", err)
	}
	if !result.Success {
		t.Fatalf("HealthCheck success=false: %s", result.Error)
	}
	if !result.Connected {
		t.Fatal("connected=false")
	}
	if result.TargetURL == "" {
		t.Error("targetUrl should not be empty")
	}
	t.Logf("connected: targetUrl=%s targetId=%s", result.TargetURL, result.TargetID)
}

// ── Phase 2: chart_get_state ─────────────────────────────────────────────────

func TestChartGetState(t *testing.T) {
	skipIfNoCDP(t)
	result, err := chart.GetState()
	if err != nil {
		t.Fatalf("chart.GetState: %v", err)
	}
	b, _ := json.Marshal(result)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)

	if success, _ := m["success"].(bool); !success {
		t.Fatalf("success=false: %v", m["error"])
	}
	if sym, _ := m["symbol"].(string); sym == "" {
		t.Error("symbol should not be empty")
	}
	if tf, _ := m["timeframe"].(string); tf == "" {
		t.Error("timeframe should not be empty")
	}
	if typ, _ := m["type"].(string); typ == "" {
		t.Error("chart type should not be empty")
	}
	t.Logf("chart state: symbol=%v timeframe=%v type=%v", m["symbol"], m["timeframe"], m["type"])
}

// ── Phase 2: quote_get ────────────────────────────────────────────────────────

func TestQuoteGet(t *testing.T) {
	skipIfNoCDP(t)
	// GetQuote("") uses the current chart symbol.
	result, err := data.GetQuote("")
	if err != nil {
		t.Fatalf("data.GetQuote: %v", err)
	}
	b, _ := json.Marshal(result)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)

	if success, _ := m["success"].(bool); !success {
		t.Fatalf("success=false: %v", m["error"])
	}
	for _, field := range []string{"symbol", "last", "close"} {
		if m[field] == nil {
			t.Errorf("quote_get: missing field %q", field)
		}
	}
	t.Logf("quote: symbol=%v last=%v", m["symbol"], m["last"])
}

// ── Phase 2: data_get_ohlcv ───────────────────────────────────────────────────

func TestDataGetOHLCV(t *testing.T) {
	skipIfNoCDP(t)
	result, err := data.GetOhlcv(5, false)
	if err != nil {
		t.Fatalf("data.GetOhlcv: %v", err)
	}
	b, _ := json.Marshal(result)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)

	if success, _ := m["success"].(bool); !success {
		t.Fatalf("success=false: %v", m["error"])
	}
	bars, _ := m["bars"].([]interface{})
	if len(bars) == 0 {
		t.Fatal("bars should not be empty")
	}
	bar0, _ := bars[0].(map[string]interface{})
	for _, field := range []string{"time", "open", "high", "low", "close", "volume"} {
		if bar0[field] == nil {
			t.Errorf("bar[0] missing field %q", field)
		}
	}
	t.Logf("ohlcv: bar_count=%v first_time=%v", m["bar_count"], bar0["time"])
}

// ── Phase 2: data_get_study_values ────────────────────────────────────────────

func TestDataGetStudyValues(t *testing.T) {
	skipIfNoCDP(t)
	result, err := data.GetStudyValues()
	if err != nil {
		t.Fatalf("data.GetStudyValues: %v", err)
	}
	b, _ := json.Marshal(result)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)

	if success, _ := m["success"].(bool); !success {
		t.Fatalf("success=false: %v", m["error"])
	}
	// studies array may be empty when no indicators are on the chart — that's valid
	studies, _ := m["studies"].([]interface{})
	t.Logf("study_count=%v studies=%d", m["study_count"], len(studies))
}

// ── Phase 2: capture_screenshot ──────────────────────────────────────────────

func TestCaptureScreenshot(t *testing.T) {
	skipIfNoCDP(t)
	result, err := capture.CaptureScreenshot("", "")
	if err != nil {
		t.Fatalf("capture.CaptureScreenshot: %v", err)
	}
	b, _ := json.Marshal(result)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)

	if success, _ := m["success"].(bool); !success {
		t.Fatalf("success=false: %v", m["error"])
	}
	hasData := m["data"] != nil && m["data"] != ""
	hasPath := m["path"] != nil && m["path"] != ""
	if !hasData && !hasPath {
		t.Error("screenshot result has neither data nor path")
	}
	if s, _ := m["data"].(string); s != "" {
		t.Logf("screenshot: base64 len=%d", len(s))
	} else {
		t.Logf("screenshot: path=%v", m["path"])
	}
}

// ── Phase 2: pine_get_source ──────────────────────────────────────────────────

func TestPineGetSource(t *testing.T) {
	skipIfNoCDP(t)
	result, err := pine.GetSource()
	if err != nil {
		t.Fatalf("pine.GetSource: %v", err)
	}
	b, _ := json.Marshal(result)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)

	if success, _ := m["success"].(bool); !success {
		// pine_get_source may legitimately fail when no Pine editor is open
		t.Logf("pine_get_source returned error (no Pine editor open?): %v", m["error"])
		return
	}
	src, _ := m["source"].(string)
	t.Logf("pine source len=%d", len(src))
}

// ── Phase 2: tv_launch shape ──────────────────────────────────────────────────

func TestHealthCheckShape(t *testing.T) {
	// Does not require CDP — verifies the result struct has required fields
	// regardless of connection state.
	result, err := health.HealthCheck()
	if err != nil {
		t.Fatalf("health.HealthCheck: %v", err)
	}
	b, _ := json.Marshal(result)
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, field := range []string{"success", "connected"} {
		if _, ok := m[field]; !ok {
			t.Errorf("missing field %q in health check result", field)
		}
	}
}
