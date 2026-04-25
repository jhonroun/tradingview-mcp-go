// Package capture implements capture_screenshot.
package capture

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jhonroun/tradingview-mcp-go/internal/cdp"
	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
)

const screenshotDir = "screenshots"

// CaptureScreenshot takes a PNG screenshot and saves it to screenshots/.
// region: "full" | "chart" | "strategy_tester"
func CaptureScreenshot(region, filename string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := os.MkdirAll(screenshotDir, 0o755); err != nil {
		return nil, fmt.Errorf("create screenshots dir: %w", err)
	}

	ts := strings.ReplaceAll(strings.ReplaceAll(time.Now().UTC().Format("2006-01-02T15-04-05"), ":", "-"), ".", "-")
	if region == "" {
		region = "full"
	}
	fname := filename
	if fname == "" {
		fname = fmt.Sprintf("tv_%s_%s", region, ts)
	}
	fname = strings.ReplaceAll(strings.ReplaceAll(fname, "/", "_"), "\\", "_")
	filePath := filepath.Join(screenshotDir, fname+".png")

	var b64data string
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		clip, err := regionClip(ctx, c, region)
		if err != nil {
			return err
		}
		if clip != nil {
			b64data, err = c.CaptureScreenshotClip(ctx, *clip)
		} else {
			b64data, err = c.CaptureScreenshot(ctx)
		}
		return err
	})
	if err != nil {
		return nil, err
	}

	raw, err := base64.StdEncoding.DecodeString(b64data)
	if err != nil {
		return nil, fmt.Errorf("decode screenshot: %w", err)
	}
	if err := os.WriteFile(filePath, raw, 0o644); err != nil {
		return nil, fmt.Errorf("write screenshot: %w", err)
	}

	return map[string]interface{}{
		"success":    true,
		"method":     "cdp",
		"file_path":  filePath,
		"region":     region,
		"size_bytes": len(raw),
	}, nil
}

// regionClip evaluates JS to get the bounding rect of a named region element.
func regionClip(ctx context.Context, c *cdp.Client, region string) (*cdp.ScreenshotClip, error) {
	var selector string
	switch region {
	case "chart":
		selector = `document.querySelector('[data-name="pane-canvas"]') || document.querySelector('[class*="chart-container"]') || document.querySelector('canvas')`
	case "strategy_tester":
		selector = `document.querySelector('[data-name="backtesting"]') || document.querySelector('[class*="strategyReport"]')`
	default:
		return nil, nil // full page
	}

	expr := fmt.Sprintf(`(function() {
		var el = %s;
		if (!el) return null;
		var r = el.getBoundingClientRect();
		return { x: r.x, y: r.y, width: r.width, height: r.height };
	})()`, selector)

	raw, err := c.Evaluate(ctx, expr, false)
	if err != nil || string(raw) == "null" {
		return nil, nil // fall back to full page
	}

	var bounds struct {
		X      float64 `json:"x"`
		Y      float64 `json:"y"`
		Width  float64 `json:"width"`
		Height float64 `json:"height"`
	}
	if err := json.Unmarshal(raw, &bounds); err != nil || bounds.Width == 0 {
		return nil, nil
	}
	return &cdp.ScreenshotClip{
		X: bounds.X, Y: bounds.Y,
		Width: bounds.Width, Height: bounds.Height,
		Scale: 1,
	}, nil
}

// RegisterTools registers capture_screenshot into the MCP registry.
func RegisterTools(reg *mcp.Registry) {
	reg.Register(mcp.ToolDef{
		Name:        "capture_screenshot",
		Description: `Take a screenshot of TradingView. region: "full" (default), "chart", or "strategy_tester". Saves to screenshots/ and returns the file path.`,
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"region":   {Type: "string", Description: `Region to capture: "full", "chart", or "strategy_tester"`},
				"filename": {Type: "string", Description: "Optional filename (without extension)"},
			},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Region   string `json:"region"`
				Filename string `json:"filename"`
			}
			_ = json.Unmarshal(args, &p)
			result, err := CaptureScreenshot(p.Region, p.Filename)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})
}
