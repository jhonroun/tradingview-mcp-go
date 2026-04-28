package pine

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/jhonroun/tradingview-mcp-go/internal/cdp"
)

const pineBackupRoot = "research/pine-source-safety"

type PineSourceSnapshot struct {
	Source       string `json:"source"`
	SourceSHA256 string `json:"source_sha256"`
	ScriptName   string `json:"script_name,omitempty"`
	ScriptType   string `json:"script_type,omitempty"`
	PineVersion  string `json:"pine_version,omitempty"`
	LineCount    int    `json:"line_count"`
	CharCount    int    `json:"char_count"`
	EditorURI    string `json:"editor_uri,omitempty"`
	LanguageID   string `json:"language_id,omitempty"`
	EditorTitle  string `json:"editor_title,omitempty"`
}

type pineBackupManifest struct {
	CreatedAt    string `json:"created_at"`
	Reason       string `json:"reason"`
	SourceSHA256 string `json:"source_sha256"`
	ScriptName   string `json:"script_name,omitempty"`
	ScriptType   string `json:"script_type,omitempty"`
	PineVersion  string `json:"pine_version,omitempty"`
	LineCount    int    `json:"line_count"`
	CharCount    int    `json:"char_count"`
	SourceFile   string `json:"source_file"`
}

type pineBackupResult struct {
	SessionDir   string `json:"session_dir"`
	ManifestPath string `json:"manifest_path"`
	SourcePath   string `json:"source_path"`
	SourceSHA256 string `json:"source_sha256"`
	ScriptName   string `json:"script_name,omitempty"`
	ScriptType   string `json:"script_type,omitempty"`
	LineCount    int    `json:"line_count"`
	CharCount    int    `json:"char_count"`
}

type loadedPineBackup struct {
	BackupPath   string
	SourcePath   string
	Source       string
	SourceSHA256 string
	Manifest     pineBackupManifest
}

type markerCountResult struct {
	ErrorCount   int
	WarningCount int
	Errors       []interface{}
	Warnings     []interface{}
}

func readSourceSnapshot(ctx context.Context, c *cdp.Client) (PineSourceSnapshot, error) {
	raw, err := c.Evaluate(ctx, `(function() {
		var m = `+findMonaco+`;
		if (!m) return null;
		var model = m.editor.getModel ? m.editor.getModel() : null;
		var title = '';
		var titleEl = document.querySelector('[data-name="pine-script-name"]')
			|| document.querySelector('[class*="scriptTitle"]')
			|| document.querySelector('[class*="pine"] [class*="title"]');
		if (titleEl) title = titleEl.textContent.trim();
		return {
			source: m.editor.getValue(),
			editor_uri: model && model.uri ? String(model.uri) : '',
			language_id: model && model.getLanguageId ? model.getLanguageId() : '',
			editor_title: title
		};
	})()`, false)
	if err != nil {
		return PineSourceSnapshot{}, err
	}
	var snapshot PineSourceSnapshot
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return PineSourceSnapshot{}, fmt.Errorf("Monaco editor getValue() returned unexpected type: %w", err)
	}
	return enrichSourceSnapshot(snapshot), nil
}

func enrichSourceSnapshot(snapshot PineSourceSnapshot) PineSourceSnapshot {
	meta := inferScriptMetadata(snapshot.Source)
	snapshot.SourceSHA256 = sourceSHA256(snapshot.Source)
	snapshot.LineCount = lineCount(snapshot.Source)
	snapshot.CharCount = len(snapshot.Source)
	if snapshot.ScriptName == "" {
		snapshot.ScriptName = meta.ScriptName
	}
	if snapshot.ScriptType == "" {
		snapshot.ScriptType = meta.ScriptType
	}
	if snapshot.PineVersion == "" {
		snapshot.PineVersion = meta.PineVersion
	}
	return snapshot
}

func snapshotToResult(snapshot PineSourceSnapshot) map[string]interface{} {
	result := map[string]interface{}{
		"source":             snapshot.Source,
		"source_sha256":      snapshot.SourceSHA256,
		"source_hash_sha256": snapshot.SourceSHA256,
		"hash":               snapshot.SourceSHA256,
		"line_count":         snapshot.LineCount,
		"char_count":         snapshot.CharCount,
	}
	if snapshot.ScriptName != "" {
		result["script_name"] = snapshot.ScriptName
	}
	if snapshot.ScriptType != "" {
		result["script_type"] = snapshot.ScriptType
	}
	if snapshot.PineVersion != "" {
		result["pine_version"] = snapshot.PineVersion
	}
	if snapshot.EditorURI != "" {
		result["editor_uri"] = snapshot.EditorURI
	}
	if snapshot.LanguageID != "" {
		result["language_id"] = snapshot.LanguageID
	}
	if snapshot.EditorTitle != "" {
		result["editor_title"] = snapshot.EditorTitle
	}
	return result
}

type scriptMetadata struct {
	ScriptName  string
	ScriptType  string
	PineVersion string
}

func inferScriptMetadata(source string) scriptMetadata {
	meta := scriptMetadata{}
	versionRe := regexp.MustCompile(`(?m)^\s*//@version\s*=\s*([0-9]+)`)
	if m := versionRe.FindStringSubmatch(source); len(m) > 1 {
		meta.PineVersion = m[1]
	}
	declRe := regexp.MustCompile(`(?m)^\s*(indicator|strategy|library|study)\s*\(\s*(?:"([^"]*)"|'([^']*)')?`)
	if m := declRe.FindStringSubmatch(source); len(m) > 0 {
		meta.ScriptType = strings.ToLower(m[1])
		if meta.ScriptType == "study" {
			meta.ScriptType = "indicator"
		}
		if len(m) > 2 && m[2] != "" {
			meta.ScriptName = m[2]
		} else if len(m) > 3 {
			meta.ScriptName = m[3]
		}
	}
	return meta
}

func sourceSHA256(source string) string {
	sum := sha256.Sum256([]byte(source))
	return fmt.Sprintf("%x", sum)
}

func lineCount(source string) int {
	if source == "" {
		return 0
	}
	return len(strings.Split(source, "\n"))
}

func setSourceInEditor(ctx context.Context, c *cdp.Client, source string) error {
	escaped, _ := json.Marshal(source)
	raw, err := c.Evaluate(ctx, `(function() {
		var m = `+findMonaco+`;
		if (!m) return false;
		m.editor.setValue(`+string(escaped)+`);
		return true;
	})()`, false)
	if err != nil {
		return err
	}
	var ok bool
	if json.Unmarshal(raw, &ok) != nil || !ok {
		return fmt.Errorf("Monaco found but setValue() failed")
	}
	return nil
}

func createPineBackup(snapshot PineSourceSnapshot, reason string) (pineBackupResult, error) {
	snapshot = enrichSourceSnapshot(snapshot)
	now := time.Now().UTC()
	dir := filepath.Join(pineBackupRoot, "session-"+now.Format("20060102T150405.000000000Z"))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return pineBackupResult{}, err
	}
	name := safeFilenamePart(snapshot.ScriptName)
	if name == "" {
		name = "pine_source"
	}
	sourcePath := filepath.Join(dir, name+".pine")
	manifestPath := filepath.Join(dir, "backup.json")
	if err := os.WriteFile(sourcePath, []byte(snapshot.Source), 0o600); err != nil {
		return pineBackupResult{}, err
	}
	manifest := pineBackupManifest{
		CreatedAt:    now.Format(time.RFC3339Nano),
		Reason:       reason,
		SourceSHA256: snapshot.SourceSHA256,
		ScriptName:   snapshot.ScriptName,
		ScriptType:   snapshot.ScriptType,
		PineVersion:  snapshot.PineVersion,
		LineCount:    snapshot.LineCount,
		CharCount:    snapshot.CharCount,
		SourceFile:   filepath.Base(sourcePath),
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return pineBackupResult{}, err
	}
	if err := os.WriteFile(manifestPath, append(data, '\n'), 0o600); err != nil {
		return pineBackupResult{}, err
	}
	return pineBackupResult{
		SessionDir:   filepath.ToSlash(dir),
		ManifestPath: filepath.ToSlash(manifestPath),
		SourcePath:   filepath.ToSlash(sourcePath),
		SourceSHA256: snapshot.SourceSHA256,
		ScriptName:   snapshot.ScriptName,
		ScriptType:   snapshot.ScriptType,
		LineCount:    snapshot.LineCount,
		CharCount:    snapshot.CharCount,
	}, nil
}

func backupToResult(backup pineBackupResult) map[string]interface{} {
	result := map[string]interface{}{
		"session_dir":   backup.SessionDir,
		"manifest_path": backup.ManifestPath,
		"source_path":   backup.SourcePath,
		"source_sha256": backup.SourceSHA256,
		"hash":          backup.SourceSHA256,
		"line_count":    backup.LineCount,
		"char_count":    backup.CharCount,
	}
	if backup.ScriptName != "" {
		result["script_name"] = backup.ScriptName
	}
	if backup.ScriptType != "" {
		result["script_type"] = backup.ScriptType
	}
	return result
}

func loadPineBackup(path, expectedSHA256 string) (loadedPineBackup, error) {
	if strings.TrimSpace(path) == "" {
		return loadedPineBackup{}, fmt.Errorf("backup_path is required")
	}
	clean := filepath.Clean(path)
	data, err := os.ReadFile(clean)
	if err != nil {
		return loadedPineBackup{}, err
	}
	backup := loadedPineBackup{BackupPath: filepath.ToSlash(clean)}
	if strings.EqualFold(filepath.Ext(clean), ".json") {
		if err := json.Unmarshal(data, &backup.Manifest); err != nil {
			return loadedPineBackup{}, fmt.Errorf("parse backup manifest: %w", err)
		}
		if backup.Manifest.SourceFile == "" {
			return loadedPineBackup{}, fmt.Errorf("backup manifest missing source_file")
		}
		sourcePath := backup.Manifest.SourceFile
		if !filepath.IsAbs(sourcePath) {
			sourcePath = filepath.Join(filepath.Dir(clean), sourcePath)
		}
		sourceBytes, err := os.ReadFile(sourcePath)
		if err != nil {
			return loadedPineBackup{}, err
		}
		backup.Source = string(sourceBytes)
		backup.SourcePath = filepath.ToSlash(sourcePath)
		if expectedSHA256 == "" {
			expectedSHA256 = backup.Manifest.SourceSHA256
		}
	} else {
		backup.Source = string(data)
		backup.SourcePath = filepath.ToSlash(clean)
	}
	if expectedSHA256 == "" {
		return loadedPineBackup{}, fmt.Errorf("expected_sha256 is required unless backup manifest contains source_sha256")
	}
	actual := sourceSHA256(backup.Source)
	if !strings.EqualFold(actual, expectedSHA256) {
		return loadedPineBackup{}, fmt.Errorf("backup SHA256 mismatch: expected %s, got %s", expectedSHA256, actual)
	}
	backup.SourceSHA256 = actual
	return backup, nil
}

func safeFilenamePart(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range value {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
		case r == '.', r == '_', r == '-':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
		if b.Len() >= 80 {
			break
		}
	}
	return strings.Trim(b.String(), "._-")
}

func getMarkers(ctx context.Context, c *cdp.Client) []interface{} {
	raw, err := c.Evaluate(ctx, getMarkersJS(), false)
	if err != nil || raw == nil {
		return []interface{}{}
	}
	var markers []interface{}
	if json.Unmarshal(raw, &markers) != nil || markers == nil {
		return []interface{}{}
	}
	return markers
}

func markerCounts(markers []interface{}) markerCountResult {
	result := markerCountResult{
		Errors:   []interface{}{},
		Warnings: []interface{}{},
	}
	for _, marker := range markers {
		m, ok := marker.(map[string]interface{})
		if !ok {
			continue
		}
		label, _ := m["severity_label"].(string)
		severity, _ := m["severity"].(float64)
		switch {
		case label == "error" || severity >= 8:
			result.ErrorCount++
			result.Errors = append(result.Errors, marker)
		case label == "warning" || severity == 4:
			result.WarningCount++
			result.Warnings = append(result.Warnings, marker)
		}
	}
	return result
}
