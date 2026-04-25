package cdp

import (
	"testing"
)

func TestFindChartTargetExact(t *testing.T) {
	targets := []Target{
		{ID: "1", Type: "page", URL: "chrome://settings"},
		{ID: "2", Type: "page", URL: "https://www.tradingview.com/chart/abc123/"},
		{ID: "3", Type: "page", URL: "https://other.com"},
	}
	got, err := FindChartTarget(targets)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "2" {
		t.Fatalf("expected target 2, got %s", got.ID)
	}
}

func TestFindChartTargetFallback(t *testing.T) {
	targets := []Target{
		{ID: "1", Type: "page", URL: "https://www.tradingview.com/"},
	}
	got, err := FindChartTarget(targets)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "1" {
		t.Fatalf("expected target 1, got %s", got.ID)
	}
}

func TestFindChartTargetNone(t *testing.T) {
	targets := []Target{
		{ID: "1", Type: "page", URL: "https://example.com"},
	}
	_, err := FindChartTarget(targets)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFindChartTargetSkipsNonPage(t *testing.T) {
	targets := []Target{
		{ID: "1", Type: "background_page", URL: "https://www.tradingview.com/chart/x/"},
		{ID: "2", Type: "page", URL: "https://www.tradingview.com/chart/y/"},
	}
	got, err := FindChartTarget(targets)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "2" {
		t.Fatalf("expected page target 2, got %s (type %s)", got.ID, got.Type)
	}
}

func TestFindChartTargetPreferChart(t *testing.T) {
	targets := []Target{
		{ID: "1", Type: "page", URL: "https://www.tradingview.com/"},
		{ID: "2", Type: "page", URL: "https://www.tradingview.com/chart/abc/"},
	}
	got, err := FindChartTarget(targets)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "2" {
		t.Fatalf("expected chart target 2, got %s", got.ID)
	}
}
