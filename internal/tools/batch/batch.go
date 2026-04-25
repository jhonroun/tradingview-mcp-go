// Package batch implements batch_run — iterate an action across symbols/timeframes.
package batch

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/jhonroun/tradingview-mcp-go/internal/cdp"
	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/capture"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/chart"
	tv "github.com/jhonroun/tradingview-mcp-go/internal/tradingview"
)

// BatchRun iterates an action across every symbol × timeframe combination.
// Mirrors batchRun() in core/batch.js.
func BatchRun(symbols, timeframes []string, action string, delayMs, ohlcvCount int) (map[string]interface{}, error) {
	if len(symbols) == 0 {
		return map[string]interface{}{"success": false, "error": "symbols array is required and must not be empty"}, nil
	}
	if delayMs <= 0 {
		delayMs = 2000
	}
	if ohlcvCount <= 0 {
		ohlcvCount = 100
	}
	if ohlcvCount > 500 {
		ohlcvCount = 500
	}
	tfs := timeframes
	if len(tfs) == 0 {
		tfs = []string{""}
	}

	results := make([]map[string]interface{}, 0, len(symbols)*len(tfs))

	for _, sym := range symbols {
		for _, tf := range tfs {
			combo := map[string]interface{}{"symbol": sym}
			if tf != "" {
				combo["timeframe"] = tf
			}

			var actionResult map[string]interface{}
			err := func() error {
				// Set symbol (includes waitForChartReady).
				if _, err := chart.SetSymbol(sym); err != nil {
					return fmt.Errorf("set symbol %s: %w", sym, err)
				}
				// Set timeframe if specified.
				if tf != "" {
					if _, err := chart.SetTimeframe(tf); err != nil {
						return fmt.Errorf("set timeframe %s: %w", tf, err)
					}
					// Extra wait after timeframe change.
					time.Sleep(500 * time.Millisecond)
				}
				// User-specified delay.
				time.Sleep(time.Duration(delayMs) * time.Millisecond)

				// Run action.
				switch action {
				case "screenshot":
					safeSym := strings.ReplaceAll(strings.ReplaceAll(sym, "/", "_"), "\\", "_")
					fname := fmt.Sprintf("batch_%s_%s", safeSym, tf)
					if tf == "" {
						fname = "batch_" + safeSym
					}
					res, err := capture.CaptureScreenshot("chart", fname)
					if err != nil {
						return fmt.Errorf("screenshot: %w", err)
					}
					if path, ok := res["file_path"].(string); ok {
						actionResult = map[string]interface{}{"file_path": filepath.ToSlash(path)}
					} else {
						actionResult = res
					}

				case "get_ohlcv":
					limit := ohlcvCount
					expr := fmt.Sprintf(`new Promise(function(resolve, reject) {
						%s.exportData({ includeTime: true, includeSeries: true, includeStudies: false })
							.then(function(result) {
								var bars = (result.data || []).slice(-%d);
								resolve({ bar_count: bars.length, last_bar: bars[bars.length - 1] || null });
							}).catch(reject);
					})`, tv.ChartAPI, limit)

					ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
					defer cancel()
					var ohlcvResult map[string]interface{}
					sessErr := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
						raw, err := c.Evaluate(ctx, expr, true)
						if err != nil {
							return err
						}
						return json.Unmarshal(raw, &ohlcvResult)
					})
					if sessErr != nil {
						return sessErr
					}
					actionResult = ohlcvResult

				case "get_strategy_results":
					time.Sleep(1 * time.Second)
					const stratExpr = `(function() {
						var metrics = {};
						var panel = document.querySelector('[data-name="backtesting"]') || document.querySelector('[class*="strategyReport"]');
						if (!panel) return { error: 'Strategy Tester not found' };
						var items = panel.querySelectorAll('[class*="reportItem"], [class*="metric"]');
						items.forEach(function(item) {
							var label = item.querySelector('[class*="label"]');
							var value = item.querySelector('[class*="value"]');
							if (label && value) metrics[label.textContent.trim()] = value.textContent.trim();
						});
						return { metric_count: Object.keys(metrics).length, metrics: metrics };
					})()`

					ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
					defer cancel()
					var stratResult map[string]interface{}
					sessErr := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
						raw, err := c.Evaluate(ctx, stratExpr, false)
						if err != nil {
							return err
						}
						return json.Unmarshal(raw, &stratResult)
					})
					if sessErr != nil {
						return sessErr
					}
					actionResult = stratResult

				default:
					return fmt.Errorf("unknown action or API not available: %s", action)
				}
				return nil
			}()

			if err != nil {
				combo["success"] = false
				combo["error"] = err.Error()
			} else {
				combo["success"] = true
				combo["result"] = actionResult
			}
			results = append(results, combo)
		}
	}

	successful := 0
	for _, r := range results {
		if s, ok := r["success"].(bool); ok && s {
			successful++
		}
	}
	return map[string]interface{}{
		"success":          true,
		"total_iterations": len(results),
		"successful":       successful,
		"failed":           len(results) - successful,
		"results":          results,
	}, nil
}

// RegisterTools registers the batch_run MCP tool.
func RegisterTools(reg *mcp.Registry) {
	reg.Register(mcp.ToolDef{
		Name:        "batch_run",
		Description: "Run an action across multiple symbols and/or timeframes",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"symbols":    {Type: "array", Description: `Array of symbols to iterate (e.g. ["BTCUSD","ETHUSD","AAPL"])`},
				"timeframes": {Type: "array", Description: `Array of timeframes (e.g. ["D","60","15"]). Optional.`},
				"action":     {Type: "string", Description: "Action to run: screenshot, get_ohlcv, get_strategy_results"},
				"delay_ms":   {Type: "number", Description: "Delay between iterations in ms (default 2000)"},
				"ohlcv_count": {Type: "number", Description: "Bar count for get_ohlcv action (default 100, max 500)"},
			},
			Required: []string{"symbols", "action"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Symbols    []string `json:"symbols"`
				Timeframes []string `json:"timeframes"`
				Action     string   `json:"action"`
				DelayMs    int      `json:"delay_ms"`
				OhlcvCount int      `json:"ohlcv_count"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return BatchRun(p.Symbols, p.Timeframes, p.Action, p.DelayMs, p.OhlcvCount)
		},
	})
}
