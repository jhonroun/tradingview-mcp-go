package cdp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

var reChartURL = regexp.MustCompile(`(?i)tradingview\.com/chart`)
var reTradingView = regexp.MustCompile(`(?i)tradingview`)

// ListTargets fetches the list of CDP targets from the debug port.
func ListTargets(ctx context.Context, host string, port int) ([]Target, error) {
	url := fmt.Sprintf("http://%s:%d/json/list", host, port)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("CDP not available at %s:%d — is TradingView running with --remote-debugging-port=%d? (%w)", host, port, port, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var targets []Target
	if err := json.Unmarshal(body, &targets); err != nil {
		return nil, fmt.Errorf("parse /json/list: %w", err)
	}
	return targets, nil
}

// FindChartTarget returns the best TradingView chart target.
// Prefers tradingview.com/chart URLs; falls back to any tradingview page.
func FindChartTarget(targets []Target) (*Target, error) {
	for i := range targets {
		t := &targets[i]
		if t.Type == "page" && reChartURL.MatchString(t.URL) {
			return t, nil
		}
	}
	for i := range targets {
		t := &targets[i]
		if t.Type == "page" && reTradingView.MatchString(t.URL) {
			return t, nil
		}
	}
	return nil, fmt.Errorf("no TradingView chart target found; is TradingView open with a chart?")
}
