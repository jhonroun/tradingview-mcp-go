package chart

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestParseStudyLimitMessageEnglish(t *testing.T) {
	text := "You applied 2 indicators - maximum available for your subscription. Current subscription: Basic, 2."
	limit, ok := parseStudyLimitMessage(text, 2)
	if !ok {
		t.Fatal("expected limit detection")
	}
	if limit != 2 {
		t.Fatalf("limit = %d, want 2", limit)
	}
}

func TestParseStudyLimitMessageRussian(t *testing.T) {
	text := "Достигнут лимит 2 индикаторов для вашей подписки."
	limit, ok := parseStudyLimitMessage(text, 2)
	if !ok {
		t.Fatal("expected Russian limit detection")
	}
	if limit != 2 {
		t.Fatalf("limit = %d, want 2", limit)
	}
}

func TestParseStudyLimitMessageIgnoresGenericError(t *testing.T) {
	_, ok := parseStudyLimitMessage("Cannot add Moving Average", 2)
	if ok {
		t.Fatal("generic add failure must not be classified as a study limit")
	}
}

func TestParseStudyLimitMessageIgnoresMaximizeText(t *testing.T) {
	_, ok := parseStudyLimitMessage("Maximize indicator pane", 2)
	if ok {
		t.Fatal("generic UI text must not be classified as a study limit")
	}
}

func TestBuildStudyLimitResultShape(t *testing.T) {
	studies := []StudyInfo{{ID: "Study_RSI_0", Name: "RSI"}, {ID: "Study_Volume_1", Name: "Volume"}}
	result := buildStudyLimitResult("add", "Moving Average", studies, studyLimitDetails{
		Message: "You applied 2 indicators - maximum available for your subscription.",
		Limit:   2,
	})

	if result["success"] != false {
		t.Fatalf("success = %v, want false", result["success"])
	}
	if result["status"] != studyLimitStatus {
		t.Fatalf("status = %v, want %s", result["status"], studyLimitStatus)
	}
	if result["limit"] != 2 {
		t.Fatalf("limit = %v, want 2", result["limit"])
	}
	if result["suggestion"] == "" {
		t.Fatal("suggestion must be populated")
	}
	current, ok := result["currentStudies"].([]StudyInfo)
	if !ok {
		t.Fatalf("currentStudies has type %T, want []StudyInfo", result["currentStudies"])
	}
	if len(current) != 2 {
		t.Fatalf("currentStudies len = %d, want 2", len(current))
	}
}

func TestSelectStudyForRemovalUsesMostRecentStudyWithEntityID(t *testing.T) {
	studies := []StudyInfo{
		{ID: "Study_RSI_0", Name: "RSI"},
		{ID: "", Name: "bad"},
		{ID: "Study_Volume_1", Name: "Volume"},
	}
	got, ok := selectStudyForRemoval(studies)
	if !ok {
		t.Fatal("expected a selected study")
	}
	if got.ID != "Study_Volume_1" {
		t.Fatalf("selected %q, want Study_Volume_1", got.ID)
	}
}

func TestAppendStudyRemovalLogToPathWritesJSONL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "research", "study-limit-detection", "removals.jsonl")
	writtenPath, err := appendStudyRemovalLogToPath(path, studyRemovalLogEntry{
		EntityID:       "Study_Volume_1",
		Name:           "Volume",
		Reason:         "study_limit_reached_allow_remove_any",
		RequestedName:  "Moving Average",
		Limit:          2,
		CurrentStudies: 2,
		AllowRemoveAny: true,
	})
	if err != nil {
		t.Fatalf("appendStudyRemovalLogToPath: %v", err)
	}
	if writtenPath == "" {
		t.Fatal("written path must be returned")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	var entry studyRemovalLogEntry
	if err := json.Unmarshal(data[:len(data)-1], &entry); err != nil {
		t.Fatalf("parse JSONL entry: %v", err)
	}
	if entry.EntityID != "Study_Volume_1" || entry.Reason == "" || entry.Timestamp == "" {
		t.Fatalf("unexpected log entry: %+v", entry)
	}
}

func TestBuildStudyAddFailedResultIsNotSilentSuccess(t *testing.T) {
	result := buildStudyAddFailedResult("Moving Average", addStudyEvaluation{
		After:     []StudyInfo{{ID: "Study_RSI_0", Name: "RSI"}},
		LimitText: "Some non-limit UI text",
	})
	if result["success"] != false {
		t.Fatalf("success = %v, want false", result["success"])
	}
	if result["status"] != "study_add_failed" {
		t.Fatalf("status = %v, want study_add_failed", result["status"])
	}
}
