package chart

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	studyLimitStatus         = "study_limit_reached"
	studyLimitRemovalLogPath = "research/study-limit-detection/removals.jsonl"
)

var studyLimitNumberRe = regexp.MustCompile(`\d+`)

type studyLimitDetails struct {
	Message string
	Limit   int
}

type studyRemovalLogEntry struct {
	Timestamp       string `json:"timestamp"`
	EntityID        string `json:"entityId"`
	Name            string `json:"name"`
	Reason          string `json:"reason"`
	RequestedName   string `json:"requestedName,omitempty"`
	Limit           int    `json:"limit,omitempty"`
	CurrentStudies  int    `json:"currentStudies,omitempty"`
	RemovalLogPath  string `json:"removalLogPath,omitempty"`
	AllowRemoveAny  bool   `json:"allowRemoveAny"`
	TradingViewPath string `json:"tradingViewPath,omitempty"`
}

func detectStudyLimit(createError, limitText string, current []StudyInfo) (studyLimitDetails, bool) {
	text := strings.TrimSpace(strings.Join(nonEmptyStrings(createError, limitText), "\n"))
	limit, ok := parseStudyLimitMessage(text, len(current))
	if !ok {
		return studyLimitDetails{}, false
	}
	return studyLimitDetails{
		Message: compactLimitMessage(text),
		Limit:   limit,
	}, true
}

func parseStudyLimitMessage(text string, currentStudies int) (int, bool) {
	lower := strings.ToLower(strings.TrimSpace(text))
	if lower == "" {
		return 0, false
	}
	hasStudyWord := containsAny(lower, []string{
		"indicator", "indicators", "study", "studies", "индикатор", "индикаторов", "исслед",
	})
	hasLimitWord := containsAny(lower, []string{
		"limit", "maximum", "subscription", "available", "applied", "reached",
		"allowed", "upgrade", "plan", "лимит", "максим", "подпис", "доступ", "план",
		"превыш", "разреш",
	})
	if !hasStudyWord || !hasLimitWord {
		return 0, false
	}

	nums := make([]int, 0)
	for _, raw := range studyLimitNumberRe.FindAllString(lower, -1) {
		n, err := strconv.Atoi(raw)
		if err == nil && n > 0 && n <= 1000 {
			nums = append(nums, n)
		}
	}
	if currentStudies > 0 {
		for _, n := range nums {
			if n == currentStudies {
				return n, true
			}
		}
	}
	for i := len(nums) - 1; i >= 0; i-- {
		return nums[i], true
	}
	return 0, true
}

func buildStudyLimitResult(action, requestedName string, current []StudyInfo, details studyLimitDetails) map[string]interface{} {
	current = normalizeStudyInfos(current)
	result := map[string]interface{}{
		"success":           false,
		"status":            studyLimitStatus,
		"error":             "TradingView study limit reached.",
		"action":            action,
		"requestedName":     requestedName,
		"currentStudies":    current,
		"currentStudyCount": len(current),
		"suggestion":        studyLimitSuggestion(),
	}
	if details.Limit > 0 {
		result["limit"] = details.Limit
	}
	if details.Message != "" {
		result["limit_message"] = details.Message
	}
	return result
}

func studyLimitSuggestion() string {
	return "Remove one study manually, upgrade the TradingView plan, or retry with allow_remove_any=true to let this tool remove the most recent study and retry."
}

func selectStudyForRemoval(studies []StudyInfo) (StudyInfo, bool) {
	for i := len(studies) - 1; i >= 0; i-- {
		study := studies[i]
		if strings.TrimSpace(study.ID) != "" {
			return study, true
		}
	}
	return StudyInfo{}, false
}

func normalizeStudyInfos(studies []StudyInfo) []StudyInfo {
	if studies == nil {
		return []StudyInfo{}
	}
	return studies
}

func appendStudyRemovalLog(entry studyRemovalLogEntry) (string, error) {
	return appendStudyRemovalLogToPath(studyLimitRemovalLogPath, entry)
}

func appendStudyRemovalLogToPath(path string, entry studyRemovalLogEntry) (string, error) {
	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}
	entry.RemovalLogPath = filepath.ToSlash(path)

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", err
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return "", err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := f.Write(append(data, '\n')); err != nil {
		return "", err
	}
	return filepath.ToSlash(path), nil
}

func containsAny(s string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(s, needle) {
			return true
		}
	}
	return false
}

func nonEmptyStrings(values ...string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func compactLimitMessage(text string) string {
	parts := strings.Fields(strings.TrimSpace(text))
	if len(parts) == 0 {
		return ""
	}
	compact := strings.Join(parts, " ")
	const max = 1200
	if len(compact) > max {
		return compact[:max] + "...(truncated)"
	}
	return compact
}
