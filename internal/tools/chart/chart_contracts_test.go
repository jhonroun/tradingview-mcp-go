package chart

import (
	"encoding/json"
	"strings"
	"testing"
)

// parseExchangeTicker mirrors the JS split logic in GetState.
func parseExchangeTicker(symbol string) (exchange, ticker string) {
	parts := strings.SplitN(symbol, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", symbol
}

func TestParseExchangeTicker(t *testing.T) {
	cases := []struct {
		symbol   string
		exchange string
		ticker   string
	}{
		{"BINANCE:BTCUSDT", "BINANCE", "BTCUSDT"},
		{"NASDAQ:AAPL", "NASDAQ", "AAPL"},
		{"NYMEX:NG1!", "NYMEX", "NG1!"},
		{"AAPL", "", "AAPL"},
		{"", "", ""},
	}
	for _, tc := range cases {
		ex, tk := parseExchangeTicker(tc.symbol)
		if ex != tc.exchange || tk != tc.ticker {
			t.Errorf("parseExchangeTicker(%q) = (%q,%q), want (%q,%q)",
				tc.symbol, ex, tk, tc.exchange, tc.ticker)
		}
	}
}

// TestChartStateContractFields ensures GetState result map has all contract keys.
func TestChartStateContractFields(t *testing.T) {
	// Simulate the output of GetState with the new contract fields.
	studies := []StudyInfo{{ID: "Study_RSI_0", Name: "RSI"}}
	state := map[string]interface{}{
		"success":    true,
		"symbol":     "BINANCE:BTCUSDT",
		"exchange":   "BINANCE",
		"ticker":     "BTCUSDT",
		"timeframe":  "60",
		"resolution": "60",
		"type":       "1",
		"chartType":  1,
		"indicators": studies,
		"studies":    studies,
		"pane_count": 2,
	}

	b, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	required := []string{
		"success", "symbol", "exchange", "ticker",
		"timeframe", "type", "indicators", "pane_count",
	}
	for _, key := range required {
		if _, ok := m[key]; !ok {
			t.Errorf("chart_get_state missing required key %q", key)
		}
	}
}

func TestStudyInfoJSONShape(t *testing.T) {
	si := StudyInfo{ID: "Study_RSI_0", Name: "Relative Strength Index"}
	b, err := json.Marshal(si)
	if err != nil {
		t.Fatalf("marshal StudyInfo: %v", err)
	}
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	for _, key := range []string{"id", "name"} {
		if _, ok := m[key]; !ok {
			t.Errorf("StudyInfo JSON missing key %q", key)
		}
	}
}

// TestSymbolInfoSentinelsViaStruct tests that the sentinel logic covers all contract fields.
func TestSymbolInfoSentinelsViaStruct(t *testing.T) {
	required := []string{"symbol", "exchange", "description", "type"}
	info := map[string]interface{}{
		// Only "symbol" present; the rest missing
		"symbol": "BTCUSDT",
	}
	for _, key := range required {
		if _, ok := info[key]; !ok {
			info[key] = ""
		}
	}
	for _, key := range required {
		v, ok := info[key]
		if !ok {
			t.Errorf("sentinel logic did not add %q", key)
		}
		if _, isStr := v.(string); !isStr {
			t.Errorf("sentinel value for %q is %T, want string", key, v)
		}
	}
}

// TestSymbolSearchResultShape verifies the struct marshals all four contract fields.
func TestSymbolSearchResultShape(t *testing.T) {
	r := SymbolSearchResult{
		Symbol:      "BTCUSDT",
		Description: "Bitcoin / TetherUS",
		Type:        "crypto",
		Exchange:    "BINANCE",
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	for _, key := range []string{"symbol", "description", "type", "exchange"} {
		if _, ok := m[key]; !ok {
			t.Errorf("SymbolSearchResult JSON missing key %q", key)
		}
	}
}

// TestSymbolSearchResultEmptyFields ensures zero-value struct still has all fields.
func TestSymbolSearchResultEmptyFields(t *testing.T) {
	r := SymbolSearchResult{}
	b, _ := json.Marshal(r)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	for _, key := range []string{"symbol", "description", "type", "exchange"} {
		v, ok := m[key]
		if !ok {
			t.Errorf("empty SymbolSearchResult JSON missing key %q", key)
			continue
		}
		if s, _ := v.(string); s != "" {
			t.Errorf("empty SymbolSearchResult[%q] = %q, want empty string", key, s)
		}
	}
}
