// Package hts implements HTS-ready composite tools for LLM integration.
// These four tools reduce multi-call round-trips by aggregating chart state,
// live price, indicator values, and futures context into single responses.
package hts

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
	charttools "github.com/jhonroun/tradingview-mcp-go/internal/tools/chart"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/data"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func strVal(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func numVal(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch vt := v.(type) {
	case float64:
		return vt
	case int:
		return float64(vt)
	case int64:
		return float64(vt)
	case string:
		s := strings.TrimSpace(strings.ReplaceAll(vt, ",", ""))
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f
		}
	}
	return 0
}

// parseFirstNumeric returns the first numeric value from a values map,
// trying float64 direct values and string representations.
func parseFirstNumeric(values map[string]interface{}) (val float64, key string, ok bool) {
	for k, v := range values {
		switch vt := v.(type) {
		case float64:
			if !math.IsNaN(vt) && !math.IsInf(vt, 0) {
				return vt, k, true
			}
		case string:
			s := strings.TrimSpace(strings.ReplaceAll(vt, ",", ""))
			if f, err := strconv.ParseFloat(s, 64); err == nil && !math.IsNaN(f) && !math.IsInf(f, 0) {
				return f, k, true
			}
		}
	}
	return 0, "", false
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

// valueDirection classifies a numeric value relative to zero.
func valueDirection(v float64) string {
	const eps = 1e-9
	if v > eps {
		return "above_zero"
	}
	if v < -eps {
		return "below_zero"
	}
	return "at_zero"
}

// studySignal returns a simple signal classification for a named indicator.
func studySignal(name string, value float64) string {
	n := strings.ToLower(name)
	switch {
	case strings.Contains(n, "rsi") || strings.Contains(n, "relative strength") || strings.Contains(n, "stoch"):
		if value >= 70 {
			return "overbought"
		}
		if value <= 30 {
			return "oversold"
		}
		return "neutral"
	case strings.Contains(n, "cci"):
		if value >= 100 {
			return "overbought"
		}
		if value <= -100 {
			return "oversold"
		}
		return "neutral"
	default:
		if value > 0 {
			return "bullish"
		}
		if value < 0 {
			return "bearish"
		}
		return "neutral"
	}
}

// ── tool functions ─────────────────────────────────────────────────────────────

// ChartContextForLLM combines chart_get_state + quote_get + top-N study values
// into a single response with a pre-built context_text string.
func ChartContextForLLM(topN int) (map[string]interface{}, error) {
	if topN <= 0 {
		topN = 5
	}

	state, err := charttools.GetState()
	if err != nil {
		return nil, fmt.Errorf("chart_get_state: %w", err)
	}

	quote, err := data.GetQuote("")
	if err != nil {
		return nil, fmt.Errorf("quote_get: %w", err)
	}

	// Best-effort: studies may be absent if no indicators are loaded.
	var indicators []interface{}
	if sv, svErr := data.GetStudyValues(); svErr == nil {
		if all, ok := sv["studies"].([]data.StudyResult); ok {
			n := topN
			if n > len(all) {
				n = len(all)
			}
			for _, s := range all[:n] {
				indicators = append(indicators, s)
			}
		}
	}

	symbol := strVal(state["symbol"])
	timeframe := strVal(state["timeframe"])
	chartType := strVal(state["type"])
	last := numVal(quote["last"])

	// Build a compact single-line context string for LLM prompt injection.
	parts := []string{
		fmt.Sprintf("Symbol: %s", symbol),
		fmt.Sprintf("TF: %s", timeframe),
		fmt.Sprintf("Price: %.4g", last),
	}
	if vol := numVal(quote["volume"]); vol > 0 {
		parts = append(parts, fmt.Sprintf("Vol: %.0f", vol))
	}
	for _, ind := range indicators {
		sr, ok := ind.(data.StudyResult)
		if !ok {
			continue
		}
		if len(sr.Plots) > 0 && sr.Plots[0].Current != nil {
			parts = append(parts, fmt.Sprintf("%s(%s): %.4g", sr.Name, sr.Plots[0].Name, *sr.Plots[0].Current))
		}
	}

	return map[string]interface{}{
		"success":         true,
		"symbol":          symbol,
		"timeframe":       timeframe,
		"chart_type":      chartType,
		"price": map[string]interface{}{
			"last":   quote["last"],
			"open":   quote["open"],
			"high":   quote["high"],
			"low":    quote["low"],
			"close":  quote["close"],
			"volume": quote["volume"],
		},
		"indicators":      indicators,
		"indicator_count": len(indicators),
		"context_text":    strings.Join(parts, " | "),
	}, nil
}

// IndicatorState finds a study by partial name match and classifies its
// current value as a direction + signal — sparing the LLM from raw arrays.
func IndicatorState(name string) (map[string]interface{}, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	sv, err := data.GetStudyValues()
	if err != nil {
		return nil, err
	}

	studies, _ := sv["studies"].([]data.StudyResult)
	nameLower := strings.ToLower(name)

	var matched *data.StudyResult
	for i := range studies {
		if strings.Contains(strings.ToLower(studies[i].Name), nameLower) {
			matched = &studies[i]
			break
		}
	}

	if matched == nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("no indicator matching %q found in current studies", name),
		}, nil
	}

	result := map[string]interface{}{
		"success":      true,
		"name":         name,
		"matched_name": matched.Name,
		"plots":        matched.Plots,
	}

	// Primary value = first plot's current value.
	if len(matched.Plots) > 0 && matched.Plots[0].Current != nil {
		primaryVal := *matched.Plots[0].Current
		result["primary_value"] = round2(primaryVal)
		result["primary_key"] = matched.Plots[0].Name
		result["direction"] = valueDirection(primaryVal)
		result["signal"] = studySignal(matched.Name, primaryVal)
		result["near_zero"] = math.Abs(primaryVal) < 0.5
	}

	return result, nil
}

// MarketSummary returns symbol, timeframe, last bar OHLCV, change%,
// volume vs 20-bar average, and all active indicator values in one call.
func MarketSummary() (map[string]interface{}, error) {
	state, err := charttools.GetState()
	if err != nil {
		return nil, fmt.Errorf("chart_get_state: %w", err)
	}

	ohlcvResult, err := data.GetOhlcv(21, false)
	if err != nil {
		return nil, fmt.Errorf("data_get_ohlcv: %w", err)
	}

	// Best-effort: indicators may be absent.
	var indicators []interface{}
	if sv, svErr := data.GetStudyValues(); svErr == nil {
		if all, ok := sv["studies"].([]data.StudyResult); ok {
			for _, s := range all {
				indicators = append(indicators, s)
			}
		}
	}

	result := map[string]interface{}{
		"success":    true,
		"symbol":     state["symbol"],
		"timeframe":  state["timeframe"],
		"chart_type": state["type"],
		"indicators": indicators,
	}

	bars, ok := ohlcvResult["bars"].([]data.Bar)
	if ok && len(bars) > 0 {
		last := bars[len(bars)-1]
		result["last_bar"] = last

		var change, changePct float64
		if len(bars) >= 2 {
			prev := bars[len(bars)-2]
			change = round2(last.Close - prev.Close)
			if prev.Close != 0 {
				changePct = round2((last.Close-prev.Close)/prev.Close*100)
			}
		}
		result["change"] = change
		result["change_pct"] = fmt.Sprintf("%.2f%%", changePct)

		// Volume vs average of prior bars (exclude last bar).
		if len(bars) >= 2 {
			prior := bars[:len(bars)-1]
			var volSum float64
			for _, b := range prior {
				volSum += b.Volume
			}
			avgVol := volSum / float64(len(prior))
			if avgVol > 0 {
				result["volume_vs_avg"] = round2(last.Volume / avgVol)
			}
		}
	}

	return result, nil
}

// ContinuousContractContext detects whether the current chart symbol is a
// continuous futures contract (e.g. NG1!, ES1!), parses its components, and
// enriches the response with symbol description and exchange from TradingView.
func ContinuousContractContext() (map[string]interface{}, error) {
	state, err := charttools.GetState()
	if err != nil {
		return nil, fmt.Errorf("chart_get_state: %w", err)
	}

	symbol := strVal(state["symbol"])

	// Strip exchange prefix: "NYMEX:NG1!" → "NG1!"
	base := symbol
	if idx := strings.LastIndex(symbol, ":"); idx >= 0 {
		base = symbol[idx+1:]
	}

	isContinuous := strings.Contains(base, "!")
	baseSymbol := base
	rollNumber := 0
	if isContinuous {
		if idx := strings.Index(base, "!"); idx > 0 {
			ch := base[idx-1]
			if ch >= '0' && ch <= '9' {
				rollNumber = int(ch - '0')
				baseSymbol = base[:idx-1]
			} else {
				baseSymbol = base[:idx]
			}
		}
	}

	result := map[string]interface{}{
		"success":       true,
		"symbol":        symbol,
		"is_continuous": isContinuous,
		"base_symbol":   baseSymbol,
		"roll_number":   rollNumber,
	}

	// Best-effort: enrich with symbolExt() data.
	if info, infoErr := charttools.SymbolInfo(); infoErr == nil {
		for _, field := range []string{"description", "exchange", "type", "currency_code", "root_description"} {
			if v, ok := info[field].(string); ok && v != "" {
				result[field] = v
			}
		}
	}

	if !isContinuous {
		result["note"] = "current symbol is not a continuous contract"
	} else {
		result["note"] = "nearest expiry and roll date are not available through the chart JS API; use TradingView's contract details panel"
	}

	return result, nil
}

// ── MCP registration ──────────────────────────────────────────────────────────

// RegisterTools adds all four HTS-ready tools to the MCP registry.
func RegisterTools(reg *mcp.Registry) {
	reg.Register(mcp.ToolDef{
		Name:        "chart_context_for_llm",
		Description: "Aggregate chart state + current price + top-N indicator values into one structured object with a pre-built context_text string for LLM prompt injection",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"top_n": {Type: "integer", Description: "Max number of indicators to include (default: 5)"},
			},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				TopN int `json:"top_n"`
			}
			_ = json.Unmarshal(args, &p)
			result, err := ChartContextForLLM(p.TopN)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "indicator_state",
		Description: "Get current value + signal direction (bullish/bearish/overbought/oversold/neutral) for a named indicator by partial name match; reduces LLM need to interpret raw value arrays",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"name": {Type: "string", Description: "Indicator name (partial, case-insensitive match against active studies)"},
			},
			Required: []string{"name"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Name string `json:"name"`
			}
			if err := json.Unmarshal(args, &p); err != nil || p.Name == "" {
				return map[string]interface{}{"success": false, "error": "name is required"}, nil
			}
			result, err := IndicatorState(p.Name)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "market_summary",
		Description: "One-call full market context: symbol, timeframe, last bar OHLCV, change%, volume vs 20-bar average, and all active indicator values",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			result, err := MarketSummary()
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "continuous_contract_context",
		Description: "For futures: detect continuous contract (NG1!, ES1!, CL2!, etc.), parse base symbol and roll number, return exchange and description from TradingView symbolExt()",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			result, err := ContinuousContractContext()
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})
}
