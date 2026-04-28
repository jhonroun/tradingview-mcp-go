package data

import (
	"math"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	SourceTradingViewUIDataWindow             = "tradingview_ui_data_window"
	SourceTradingViewUIDOM                    = "tradingview_ui_dom"
	SourceTradingViewStudyModel               = "tradingview_study_model"
	ReliabilityDisplayValueLocalizedUIString  = "display_value_localized_ui_string"
	ReliabilityPineRuntimeUnstableInternal    = "reliable_pine_runtime_value_unstable_internal_path"
	reliableForTradingLogicFromDisplayStrings = false
	reliableForTradingLogicFromStudyModel     = true
)

// ParseDisplayNumber parses localized TradingView display strings. It is only
// for UI/Data Window fallbacks; numeric chart/study model values should stay as
// their original float values.
func ParseDisplayNumber(input string) (float64, bool) {
	s := strings.TrimSpace(input)
	if s == "" {
		return 0, false
	}

	s = strings.NewReplacer(
		"\u00a0", " ",
		"\u202f", " ",
		"−", "-",
		"–", "-",
	).Replace(s)
	s = strings.TrimSpace(s)
	lower := strings.ToLower(s)
	switch lower {
	case "", "na", "n/a", "null", "none", "∅", "—", "-", "--":
		return 0, false
	}

	negative := false
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		negative = true
		s = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(s, "("), ")"))
	}
	if strings.HasPrefix(s, "+") {
		s = strings.TrimSpace(strings.TrimPrefix(s, "+"))
	}
	if strings.HasPrefix(s, "-") {
		negative = !negative
		s = strings.TrimSpace(strings.TrimPrefix(s, "-"))
	}

	multiplier := 1.0
	s = strings.TrimSpace(strings.TrimSuffix(s, "%"))
	if r, size := lastRune(s); size > 0 {
		switch unicode.ToUpper(r) {
		case 'K':
			multiplier = 1_000
			s = strings.TrimSpace(s[:len(s)-size])
		case 'M':
			multiplier = 1_000_000
			s = strings.TrimSpace(s[:len(s)-size])
		case 'B':
			multiplier = 1_000_000_000
			s = strings.TrimSpace(s[:len(s)-size])
		case 'T':
			multiplier = 1_000_000_000_000
			s = strings.TrimSpace(s[:len(s)-size])
		}
	}

	var b strings.Builder
	hasDigit := false
	for _, r := range s {
		switch {
		case unicode.IsDigit(r):
			hasDigit = true
			b.WriteRune(r)
		case r == ',' || r == '.':
			b.WriteRune(r)
		case r == '\'' || unicode.IsSpace(r):
			// thousands separators
		}
	}
	if !hasDigit {
		return 0, false
	}

	normalized := normalizeDisplaySeparators(b.String())
	if normalized == "" {
		return 0, false
	}
	v, err := strconv.ParseFloat(normalized, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, false
	}
	if negative {
		v = -v
	}
	return v * multiplier, true
}

func lastRune(s string) (rune, int) {
	if s == "" {
		return 0, 0
	}
	return utf8.DecodeLastRuneInString(s)
}

func normalizeDisplaySeparators(s string) string {
	if s == "" {
		return ""
	}
	lastComma := strings.LastIndex(s, ",")
	lastDot := strings.LastIndex(s, ".")

	switch {
	case lastComma >= 0 && lastDot >= 0:
		if lastComma > lastDot {
			return decimalAt(s, lastComma)
		}
		return decimalAt(s, lastDot)
	case lastComma >= 0:
		return normalizeSingleSeparator(s, ",")
	case lastDot >= 0:
		return normalizeSingleSeparator(s, ".")
	default:
		return s
	}
}

func normalizeSingleSeparator(s, sep string) string {
	parts := strings.Split(s, sep)
	if len(parts) == 1 {
		return s
	}
	if len(parts) > 2 && allThreeDigitGroups(parts[1:]) {
		return strings.Join(parts, "")
	}
	last := strings.LastIndex(s, sep)
	after := len(s) - last - len(sep)
	if sep == "," && len(parts) == 2 && after == 3 {
		return strings.ReplaceAll(s, sep, "")
	}
	return decimalAt(s, last)
}

func allThreeDigitGroups(groups []string) bool {
	if len(groups) == 0 {
		return false
	}
	for _, g := range groups {
		if len(g) != 3 {
			return false
		}
	}
	return true
}

func decimalAt(s string, idx int) string {
	var b strings.Builder
	for i, r := range s {
		if r != ',' && r != '.' {
			b.WriteRune(r)
			continue
		}
		if i == idx {
			b.WriteByte('.')
		}
	}
	return b.String()
}
