package pine

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Diagnostic is a single static analysis finding.
type Diagnostic struct {
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

// Analyze runs offline static analysis on Pine Script source and returns
// structured diagnostics. No CDP connection is required.
func Analyze(source string) map[string]interface{} {
	lines := strings.Split(source, "\n")
	var diagnostics []Diagnostic

	// Detect Pine version.
	isV6 := false
	versionNum := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//@version=6") {
			isV6 = true
			versionNum = 6
			break
		}
		if strings.HasPrefix(trimmed, "//@version=") {
			if v, err := strconv.Atoi(strings.TrimPrefix(trimmed, "//@version=")); err == nil {
				versionNum = v
			}
			break
		}
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}
		break
	}
	_ = isV6

	// Track array declarations: name → {size, line}.
	type arrayInfo struct {
		size int  // -1 means unknown
		line int
	}
	arrays := make(map[string]arrayInfo)

	reArrayFrom := regexp.MustCompile(`(\w+)\s*=\s*array\.from\(([^)]*)\)`)
	reArrayNew := regexp.MustCompile(`(\w+)\s*=\s*array\.new(?:<\w+>|_\w+)?\((\d+)?`)

	for i, line := range lines {
		if m := reArrayFrom.FindStringSubmatch(line); m != nil {
			name := strings.TrimSpace(m[1])
			args := strings.TrimSpace(m[2])
			size := 0
			if args != "" {
				size = len(strings.Split(args, ","))
			}
			arrays[name] = arrayInfo{size: size, line: i + 1}
			continue
		}
		if m := reArrayNew.FindStringSubmatch(line); m != nil {
			name := strings.TrimSpace(m[1])
			size := -1
			if m[2] != "" {
				if n, err := strconv.Atoi(m[2]); err == nil {
					size = n
				}
			}
			arrays[name] = arrayInfo{size: size, line: i + 1}
		}
	}

	// Check array.get/set for literal out-of-bounds indices.
	reGetSet := regexp.MustCompile(`array\.(get|set)\(\s*(\w+)\s*,\s*(-?\d+)`)
	for i, line := range lines {
		for _, m := range reGetSet.FindAllStringSubmatchIndex(line, -1) {
			method := line[m[2]:m[3]]
			arrName := line[m[4]:m[5]]
			idx, _ := strconv.Atoi(line[m[6]:m[7]])
			info, ok := arrays[arrName]
			if !ok || info.size < 0 {
				continue
			}
			if idx < 0 || idx >= info.size {
				diagnostics = append(diagnostics, Diagnostic{
					Line:   i + 1,
					Column: m[0] + 1,
					Message: fmt.Sprintf("array.%s(%s, %d) — index %d out of bounds (array size is %d)",
						method, arrName, idx, idx, info.size),
					Severity: "error",
				})
			}
		}
	}

	// Check array.first()/last() on zero-size arrays.
	reFirstLast := regexp.MustCompile(`(\w+)\.(first|last)\(\)`)
	for i, line := range lines {
		for _, m := range reFirstLast.FindAllStringSubmatchIndex(line, -1) {
			arrName := line[m[2]:m[3]]
			method := line[m[4]:m[5]]
			if arrName == "array" {
				continue
			}
			info, ok := arrays[arrName]
			if ok && info.size == 0 {
				diagnostics = append(diagnostics, Diagnostic{
					Line:   i + 1,
					Column: m[0] + 1,
					Message: fmt.Sprintf("%s.%s() called on possibly empty array (declared with size 0)",
						arrName, method),
					Severity: "warning",
				})
			}
		}
	}

	// Check strategy.entry/close without strategy() declaration.
	hasStrategyDecl := false
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "strategy(") {
			hasStrategyDecl = true
			break
		}
	}
	if !hasStrategyDecl {
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.Contains(trimmed, "strategy.entry") || strings.Contains(trimmed, "strategy.close") {
				diagnostics = append(diagnostics, Diagnostic{
					Line:     i + 1,
					Column:   1,
					Message:  "strategy.entry/close used but no strategy() declaration found — did you mean to use indicator()?",
					Severity: "error",
				})
				break
			}
		}
	}

	// Warn about old Pine versions (< 5).
	if versionNum > 0 && versionNum < 5 {
		diagnostics = append(diagnostics, Diagnostic{
			Line:     1,
			Column:   1,
			Message:  fmt.Sprintf("Script uses Pine v%d — consider upgrading to v6 for latest features", versionNum),
			Severity: "info",
		})
	}

	result := map[string]interface{}{
		"success":     true,
		"issue_count": len(diagnostics),
		"diagnostics": diagnostics,
	}
	if len(diagnostics) == 0 {
		result["note"] = "No static analysis issues found. Use pine_compile or pine_smart_compile for full server-side compilation check."
	}
	return result
}
