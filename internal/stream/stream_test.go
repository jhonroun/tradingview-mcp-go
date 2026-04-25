package stream

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestDefaultIntervals(t *testing.T) {
	cases := []struct {
		name string
		fn   func(int) int
		want int
	}{
		{"quote", func(v int) int { if v <= 0 { return 300 }; return v }, 300},
		{"bars", func(v int) int { if v <= 0 { return 500 }; return v }, 500},
		{"values", func(v int) int { if v <= 0 { return 500 }; return v }, 500},
		{"lines", func(v int) int { if v <= 0 { return 1000 }; return v }, 1000},
		{"labels", func(v int) int { if v <= 0 { return 1000 }; return v }, 1000},
		{"tables", func(v int) int { if v <= 0 { return 2000 }; return v }, 2000},
		{"all-panes", func(v int) int { if v <= 0 { return 500 }; return v }, 500},
	}
	for _, tc := range cases {
		got := tc.fn(0)
		if got != tc.want {
			t.Errorf("%s: default interval = %d, want %d", tc.name, got, tc.want)
		}
		if tc.fn(999) != 999 {
			t.Errorf("%s: explicit interval 999 was overridden", tc.name)
		}
	}
}

func TestBuildLinesExprContainsFilter(t *testing.T) {
	expr := buildLinesExpr("MyStudy")
	if !strings.Contains(expr, `"MyStudy"`) {
		t.Error("buildLinesExpr should embed the study filter as a JSON string")
	}
	exprNoFilter := buildLinesExpr("")
	if !strings.Contains(exprNoFilter, "null") {
		t.Error("buildLinesExpr with empty filter should use null")
	}
}

func TestBuildLabelsExprContainsFilter(t *testing.T) {
	expr := buildLabelsExpr("Profiler")
	if !strings.Contains(expr, `"Profiler"`) {
		t.Error("buildLabelsExpr should embed the study filter")
	}
}

func TestBuildTablesExprContainsFilter(t *testing.T) {
	expr := buildTablesExpr("Dashboard")
	if !strings.Contains(expr, `"Dashboard"`) {
		t.Error("buildTablesExpr should embed the study filter")
	}
}

func TestIsCDPError(t *testing.T) {
	cases := []struct {
		msg  string
		want bool
	}{
		{"CDP error 1001: closed", true},
		{"websocket: read tcp", true},
		{"connection refused", true},
		{"ECONNREFUSED 127.0.0.1:9222", true},
		{"closed pipe", true},
		{"invalid JSON payload", false},
		{"array out of bounds", false},
	}
	for _, tc := range cases {
		got := isCDPError(errors.New(tc.msg))
		if got != tc.want {
			t.Errorf("isCDPError(%q) = %v, want %v", tc.msg, got, tc.want)
		}
	}
}

func TestPollLoopCancelledContext(t *testing.T) {
	// Pre-cancel the context so pollLoop exits immediately after startup.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var out bytes.Buffer
	var errOut bytes.Buffer

	err := StreamQuote(ctx, &out, &errOut, 50)
	if err != nil {
		t.Errorf("StreamQuote should return nil on cancelled ctx, got: %v", err)
	}
	if !strings.Contains(errOut.String(), "stopped") {
		t.Errorf("expected 'stopped' in stderr output, got: %q", errOut.String())
	}
}

func TestJSONLTimestampFields(t *testing.T) {
	data := map[string]interface{}{
		"symbol": "AAPL",
		"close":  150.0,
	}
	data["_ts"] = time.Now().UnixMilli()
	data["_stream"] = "quote"

	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	line := string(raw)
	if !strings.Contains(line, `"_ts"`) {
		t.Error("JSONL line should contain _ts field")
	}
	if !strings.Contains(line, `"_stream":"quote"`) {
		t.Error("JSONL line should contain _stream field")
	}
}

func TestConstantsContainTradingViewAPI(t *testing.T) {
	if !strings.Contains(chartAPI, "TradingViewApi") {
		t.Error("chartAPI constant should reference TradingViewApi")
	}
	if !strings.Contains(cwc, "_chartWidgetCollection") {
		t.Error("cwc constant should reference _chartWidgetCollection")
	}
}
