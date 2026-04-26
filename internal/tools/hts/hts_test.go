package hts

import (
	"math"
	"testing"

	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
)

// ── tool registration ─────────────────────────────────────────────────────────

func TestRegisterHTSToolNames(t *testing.T) {
	reg := mcp.NewRegistry()
	RegisterTools(reg)
	tools := reg.List()

	want := []string{
		"chart_context_for_llm",
		"indicator_state",
		"market_summary",
		"continuous_contract_context",
	}
	if len(tools) != len(want) {
		t.Fatalf("RegisterTools registered %d tools, want %d", len(tools), len(want))
	}
	for i, w := range want {
		if tools[i].Name != w {
			t.Errorf("tool[%d] = %q, want %q", i, tools[i].Name, w)
		}
		if tools[i].Description == "" {
			t.Errorf("tool[%d] %q has empty description", i, w)
		}
		if tools[i].InputSchema.Type == "" {
			t.Errorf("tool[%d] %q has empty inputSchema.type", i, w)
		}
	}
}

// ── parseFirstNumeric ─────────────────────────────────────────────────────────

func TestParseFirstNumericFloat(t *testing.T) {
	f, key, ok := parseFirstNumeric(map[string]interface{}{"RSI": float64(65.3)})
	if !ok || key != "RSI" || math.Abs(f-65.3) > 1e-9 {
		t.Errorf("expected RSI=65.3, got key=%q f=%v ok=%v", key, f, ok)
	}
}

func TestParseFirstNumericString(t *testing.T) {
	f, _, ok := parseFirstNumeric(map[string]interface{}{"MACD": "1.234"})
	if !ok || math.Abs(f-1.234) > 1e-9 {
		t.Errorf("expected 1.234 from string, got %v ok=%v", f, ok)
	}
}

func TestParseFirstNumericStringWithComma(t *testing.T) {
	f, _, ok := parseFirstNumeric(map[string]interface{}{"Price": "1,234.56"})
	if !ok || math.Abs(f-1234.56) > 1e-9 {
		t.Errorf("expected 1234.56, got %v ok=%v", f, ok)
	}
}

func TestParseFirstNumericNonNumericString(t *testing.T) {
	_, _, ok := parseFirstNumeric(map[string]interface{}{"val": "n/a"})
	if ok {
		t.Error("expected ok=false for non-numeric string")
	}
}

func TestParseFirstNumericEmpty(t *testing.T) {
	_, _, ok := parseFirstNumeric(map[string]interface{}{})
	if ok {
		t.Error("expected ok=false for empty map")
	}
}

// ── valueDirection ────────────────────────────────────────────────────────────

func TestValueDirection(t *testing.T) {
	cases := []struct {
		v    float64
		want string
	}{
		{1.0, "above_zero"},
		{-1.0, "below_zero"},
		{0.0, "at_zero"},
		{1e-10, "at_zero"},  // within eps
		{-1e-10, "at_zero"}, // within eps
		{2e-9, "above_zero"},
	}
	for _, tc := range cases {
		if got := valueDirection(tc.v); got != tc.want {
			t.Errorf("valueDirection(%v) = %q, want %q", tc.v, got, tc.want)
		}
	}
}

// ── studySignal ───────────────────────────────────────────────────────────────

func TestStudySignalRSI(t *testing.T) {
	cases := []struct {
		name  string
		value float64
		want  string
	}{
		{"RSI", 75, "overbought"},
		{"RSI", 25, "oversold"},
		{"RSI", 50, "neutral"},
		{"Relative Strength Index", 80, "overbought"},
		{"Stochastic", 85, "overbought"},
		{"Stochastic %K", 20, "oversold"},
	}
	for _, tc := range cases {
		if got := studySignal(tc.name, tc.value); got != tc.want {
			t.Errorf("studySignal(%q, %v) = %q, want %q", tc.name, tc.value, got, tc.want)
		}
	}
}

func TestStudySignalCCI(t *testing.T) {
	cases := []struct {
		value float64
		want  string
	}{
		{150, "overbought"},
		{-150, "oversold"},
		{50, "neutral"},
	}
	for _, tc := range cases {
		if got := studySignal("CCI", tc.value); got != tc.want {
			t.Errorf("studySignal(CCI, %v) = %q, want %q", tc.value, got, tc.want)
		}
	}
}

func TestStudySignalGeneric(t *testing.T) {
	if got := studySignal("MACD", 1.5); got != "bullish" {
		t.Errorf("MACD positive should be bullish, got %q", got)
	}
	if got := studySignal("MACD", -0.5); got != "bearish" {
		t.Errorf("MACD negative should be bearish, got %q", got)
	}
	if got := studySignal("EMA", 0); got != "neutral" {
		t.Errorf("zero value should be neutral, got %q", got)
	}
}

// ── strVal / numVal ───────────────────────────────────────────────────────────

func TestStrVal(t *testing.T) {
	if strVal(nil) != "" {
		t.Error("strVal(nil) should return empty string")
	}
	if strVal("hello") != "hello" {
		t.Error("strVal string passthrough failed")
	}
}

func TestNumVal(t *testing.T) {
	if numVal(nil) != 0 {
		t.Error("numVal(nil) should return 0")
	}
	if numVal(float64(3.14)) != 3.14 {
		t.Error("numVal float64 passthrough failed")
	}
	if numVal("2.71") != 2.71 {
		t.Errorf("numVal string parse: got %v, want 2.71", numVal("2.71"))
	}
	if numVal("bad") != 0 {
		t.Error("numVal bad string should return 0")
	}
}

// ── continuous contract parsing ───────────────────────────────────────────────

func TestContinuousContractSymbolParsing(t *testing.T) {
	cases := []struct {
		symbol         string
		wantContinuous bool
		wantBase       string
		wantRoll       int
	}{
		{"NG1!", true, "NG", 1},
		{"ES1!", true, "ES", 1},
		{"CL2!", true, "CL", 2},
		{"NQ1!", true, "NQ", 1},
		{"AAPL", false, "AAPL", 0},
		{"MSFT", false, "MSFT", 0},
	}

	for _, tc := range cases {
		base := tc.symbol
		if idx := lastIndex(base, ":"); idx >= 0 {
			base = base[idx+1:]
		}
		isCont := contains(base, "!")
		if isCont != tc.wantContinuous {
			t.Errorf("symbol %q isContinuous=%v, want %v", tc.symbol, isCont, tc.wantContinuous)
			continue
		}
		if !isCont {
			continue
		}
		baseSymbol := base
		rollNum := 0
		if idx := indexRune(base, '!'); idx > 0 {
			ch := base[idx-1]
			if ch >= '0' && ch <= '9' {
				rollNum = int(ch - '0')
				baseSymbol = base[:idx-1]
			} else {
				baseSymbol = base[:idx]
			}
		}
		if baseSymbol != tc.wantBase {
			t.Errorf("symbol %q base=%q, want %q", tc.symbol, baseSymbol, tc.wantBase)
		}
		if rollNum != tc.wantRoll {
			t.Errorf("symbol %q roll=%d, want %d", tc.symbol, rollNum, tc.wantRoll)
		}
	}
}

func TestContinuousContractWithExchangePrefix(t *testing.T) {
	symbol := "NYMEX:NG1!"
	base := symbol
	if idx := lastIndex(base, ":"); idx >= 0 {
		base = base[idx+1:]
	}
	if !contains(base, "!") {
		t.Error("NYMEX:NG1! should be detected as continuous")
	}
}

// helpers mirroring the production logic to keep tests package-local
func lastIndex(s, sub string) int {
	idx := -1
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			idx = i
		}
	}
	return idx
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func indexRune(s string, r rune) int {
	for i, c := range s {
		if c == r {
			return i
		}
	}
	return -1
}

// ── round2 ────────────────────────────────────────────────────────────────────

func TestRound2(t *testing.T) {
	cases := []struct {
		input float64
		want  float64
	}{
		{1.2345, 1.23},
		{1.235, 1.24},
		{-0.001, 0.0},
		{100.999, 101.0},
	}
	for _, tc := range cases {
		if got := round2(tc.input); got != tc.want {
			t.Errorf("round2(%v) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

// ── indicator_state missing name guard ───────────────────────────────────────

func TestIndicatorStateMissingName(t *testing.T) {
	reg := mcp.NewRegistry()
	RegisterTools(reg)
	result, err := reg.Call("indicator_state", []byte(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected result type: %T", result)
	}
	if m["success"] != false {
		t.Error("expected success=false for missing name")
	}
}
