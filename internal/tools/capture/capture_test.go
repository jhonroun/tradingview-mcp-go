package capture

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestScreenshotFilePathNormalizesPNGExtension(t *testing.T) {
	tests := []struct {
		name     string
		region   string
		filename string
		ts       string
		want     string
	}{
		{
			name:     "adds png for plain filename",
			region:   "chart",
			filename: "session",
			ts:       "2026-04-27T00-00-00",
			want:     filepath.Join(screenshotDir, "session.png"),
		},
		{
			name:     "keeps png filename",
			region:   "chart",
			filename: "session.png",
			ts:       "2026-04-27T00-00-00",
			want:     filepath.Join(screenshotDir, "session.png"),
		},
		{
			name:     "keeps uppercase png filename",
			region:   "chart",
			filename: "session.PNG",
			ts:       "2026-04-27T00-00-00",
			want:     filepath.Join(screenshotDir, "session.PNG"),
		},
		{
			name:     "adds png for path without extension",
			region:   "chart",
			filename: "nested/path/session",
			ts:       "2026-04-27T00-00-00",
			want:     filepath.Join(screenshotDir, "nested_path_session.png"),
		},
		{
			name:     "adds png for windows path without extension",
			region:   "chart",
			filename: `nested\path\session`,
			ts:       "2026-04-27T00-00-00",
			want:     filepath.Join(screenshotDir, "nested_path_session.png"),
		},
		{
			name:     "default filename",
			region:   "chart",
			filename: "",
			ts:       "2026-04-27T00-00-00",
			want:     filepath.Join(screenshotDir, "tv_chart_2026-04-27T00-00-00.png"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := screenshotFilePath(tt.region, tt.filename, tt.ts)
			if err != nil {
				t.Fatalf("screenshotFilePath() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("screenshotFilePath() = %q, want %q", got, tt.want)
			}
			if strings.Contains(strings.ToLower(got), ".png.png") {
				t.Fatalf("screenshotFilePath() contains duplicate png extension: %q", got)
			}
		})
	}
}

func TestScreenshotFilePathRejectsNonPNGExtension(t *testing.T) {
	tests := []string{
		"session.jpg",
		"session.webp",
		"nested/path/session.jpeg",
		`nested\path\session.gif`,
	}

	for _, filename := range tests {
		t.Run(filename, func(t *testing.T) {
			got, err := screenshotFilePath("chart", filename, "2026-04-27T00-00-00")
			if err == nil {
				t.Fatalf("screenshotFilePath() error = nil, got path %q", got)
			}
			if !strings.Contains(err.Error(), ".png") {
				t.Fatalf("screenshotFilePath() error = %q, want .png guidance", err.Error())
			}
		})
	}
}
