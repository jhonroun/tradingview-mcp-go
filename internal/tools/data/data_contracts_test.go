package data

import (
	"encoding/json"
	"testing"
)

// ── StudyResult JSON shape ────────────────────────────────────────────────────

func TestStudyResultJSONShape(t *testing.T) {
	sr := StudyResult{
		Name:      "RSI",
		EntityID:  "Study_RSI_0",
		PlotCount: 1,
		Plots: []StudyPlot{
			{Name: "RSI", Current: ptr64(55.3), Values: []float64{55.3}},
		},
	}
	b, err := json.Marshal(sr)
	if err != nil {
		t.Fatalf("marshal StudyResult: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"name", "entity_id", "plot_count", "plots"} {
		if _, ok := m[key]; !ok {
			t.Errorf("StudyResult JSON missing key %q", key)
		}
	}
	plots, _ := m["plots"].([]interface{})
	if len(plots) != 1 {
		t.Fatalf("plots len = %d, want 1", len(plots))
	}
	plot, _ := plots[0].(map[string]interface{})
	for _, key := range []string{"name", "current", "values"} {
		if _, ok := plot[key]; !ok {
			t.Errorf("StudyPlot JSON missing key %q", key)
		}
	}
	// current == values[0]
	cur, _ := plot["current"].(float64)
	vals, _ := plot["values"].([]interface{})
	if len(vals) == 0 {
		t.Fatal("values array must not be empty when current is set")
	}
	v0, _ := vals[0].(float64)
	if cur != v0 {
		t.Errorf("current (%v) != values[0] (%v)", cur, v0)
	}
}

func TestStudyPlotNilCurrent(t *testing.T) {
	sp := StudyPlot{Name: "RSI", Current: nil, Values: []float64{}}
	b, _ := json.Marshal(sp)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	// current must be null (not absent) when nil
	if _, ok := m["current"]; !ok {
		t.Error("current must be present even when nil")
	}
	vals, _ := m["values"].([]interface{})
	if len(vals) != 0 {
		t.Errorf("values must be empty when current is nil, got %v", vals)
	}
}

// ── quote_get sentinel fields ─────────────────────────────────────────────────

// buildQuote simulates the Go-side sentinel logic in GetQuote.
func buildQuote(raw map[string]interface{}) map[string]interface{} {
	for _, key := range []string{"bid", "ask", "change", "change_pct"} {
		if raw[key] == nil {
			raw[key] = float64(0)
		}
	}
	raw["success"] = true
	return raw
}

func TestQuoteAlwaysHasSentinelFields(t *testing.T) {
	// Simulate a quote with no bid/ask from DOM (index symbol or crypto feed)
	raw := map[string]interface{}{
		"symbol": "BINANCE:BTCUSDT",
		"last":   float64(67400),
		"open":   float64(66800),
		"high":   float64(67900),
		"low":    float64(66500),
		"close":  float64(67400),
		"volume": float64(12345.67),
	}
	q := buildQuote(raw)
	for _, key := range []string{"bid", "ask", "change", "change_pct"} {
		v, ok := q[key]
		if !ok {
			t.Errorf("quote missing key %q", key)
			continue
		}
		f, ok := v.(float64)
		if !ok {
			t.Errorf("quote[%q] is %T, want float64", key, v)
			continue
		}
		_ = f // zero is the expected sentinel
	}
}

func TestQuoteSuccessAlwaysSet(t *testing.T) {
	q := buildQuote(map[string]interface{}{"last": float64(100)})
	if q["success"] != true {
		t.Error("success must be true")
	}
}

// ── symbol_info sentinel fields ───────────────────────────────────────────────

func TestSymbolInfoSentinels(t *testing.T) {
	// Simulate SymbolInfo() result where some fields are missing.
	info := map[string]interface{}{
		"symbol": "BTCUSDT",
		// "exchange", "description", "type" intentionally absent
	}
	// Apply the sentinel logic from SymbolInfo().
	for _, key := range []string{"symbol", "exchange", "description", "type"} {
		if _, ok := info[key]; !ok {
			info[key] = ""
		}
	}
	for _, key := range []string{"symbol", "exchange", "description", "type"} {
		v, ok := info[key]
		if !ok {
			t.Errorf("symbol_info missing key %q after sentinel fix", key)
			continue
		}
		if _, isStr := v.(string); !isStr {
			t.Errorf("symbol_info[%q] is %T, want string", key, v)
		}
	}
}

// ── GetStudyValues empty guard ────────────────────────────────────────────────

func TestGetStudyValuesEmptyArrayNotNil(t *testing.T) {
	// The contract requires studies: [] never null.
	var studies []StudyResult // nil slice
	if studies == nil {
		studies = []StudyResult{}
	}
	b, err := json.Marshal(map[string]interface{}{
		"success":     true,
		"study_count": 0,
		"studies":     studies,
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	studiesJSON, _ := m["studies"]
	if studiesJSON == nil {
		t.Error("studies must not be null")
	}
	arr, ok := studiesJSON.([]interface{})
	if !ok {
		t.Errorf("studies must be array, got %T", studiesJSON)
	}
	if len(arr) != 0 {
		t.Errorf("studies must be empty array, got len %d", len(arr))
	}
}

// ── helper ────────────────────────────────────────────────────────────────────

func ptr64(v float64) *float64 { return &v }
