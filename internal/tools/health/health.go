// Package health implements tv_health_check, tv_discover, tv_ui_state, and tv_launch.
package health

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jhonroun/tradingview-mcp-go/internal/cdp"
	"github.com/jhonroun/tradingview-mcp-go/internal/launcher"
	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
)

// HealthCheckResult mirrors the Node.js healthCheck() response shape.
type HealthCheckResult struct {
	Success   bool   `json:"success"`
	Connected bool   `json:"connected"`
	TargetURL string `json:"targetUrl,omitempty"`
	TargetID  string `json:"targetId,omitempty"`
	Error     string `json:"error,omitempty"`
	Hint      string `json:"hint,omitempty"`
}

// LaunchArgs are the optional arguments for tv_launch / Launch().
type LaunchArgs struct {
	Port         *int    `json:"port,omitempty"`
	KillExisting *bool   `json:"kill_existing,omitempty"`
	TvPath       *string `json:"tv_path,omitempty"`
}

// HealthCheck verifies CDP connectivity and returns chart target info.
func HealthCheck() (*HealthCheckResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	targets, err := cdp.ListTargets(ctx, "localhost", 9222)
	if err != nil {
		return &HealthCheckResult{
			Success:   false,
			Connected: false,
			Error:     err.Error(),
			Hint:      "TradingView is not running with CDP enabled. Use the tv_launch tool to start it automatically.",
		}, nil
	}
	target, err := cdp.FindChartTarget(targets)
	if err != nil {
		return &HealthCheckResult{
			Success:   false,
			Connected: true,
			Error:     err.Error(),
			Hint:      "TradingView is running but no chart is open. Open a chart to enable chart tools.",
		}, nil
	}
	return &HealthCheckResult{
		Success:   true,
		Connected: true,
		TargetURL: target.URL,
		TargetID:  target.ID,
	}, nil
}

// Discover reports which known TradingView API paths are reachable.
func Discover() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	targets, err := cdp.ListTargets(ctx, "localhost", 9222)
	if err != nil {
		return nil, fmt.Errorf("CDP not available: %w", err)
	}
	target, err := cdp.FindChartTarget(targets)
	if err != nil {
		return nil, err
	}
	client, err := cdp.Connect(ctx, target)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	if err := client.EnableDomains(ctx); err != nil {
		return nil, err
	}

	knownPaths := map[string]string{
		"chartApi":              "window.TradingViewApi._activeChartWidgetWV.value()",
		"chartWidgetCollection": "window.TradingViewApi._chartWidgetCollection",
		"replayApi":             "window.TradingViewApi._replayApi",
		"alertService":          "window.TradingViewApi._alertService",
	}
	available := make(map[string]bool, len(knownPaths))
	for name, expr := range knownPaths {
		check := fmt.Sprintf("typeof (%s) !== 'undefined' && (%s) !== null", expr, expr)
		val, err := client.Evaluate(ctx, check, false)
		if err == nil {
			var b bool
			_ = json.Unmarshal(val, &b)
			available[name] = b
		} else {
			available[name] = false
		}
	}
	return map[string]interface{}{"success": true, "paths": available}, nil
}

// UIState returns currently visible UI panels via DOM inspection.
func UIState() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	targets, err := cdp.ListTargets(ctx, "localhost", 9222)
	if err != nil {
		return nil, fmt.Errorf("CDP not available: %w", err)
	}
	target, err := cdp.FindChartTarget(targets)
	if err != nil {
		return nil, err
	}
	client, err := cdp.Connect(ctx, target)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	if err := client.EnableDomains(ctx); err != nil {
		return nil, err
	}

	const expr = `(function() {
		var panels = [];
		document.querySelectorAll('[data-name]').forEach(function(el) {
			if (el.offsetParent !== null) panels.push(el.getAttribute('data-name'));
		});
		return JSON.stringify({success: true, visiblePanels: panels.filter(function(v,i,a){return a.indexOf(v)===i;})});
	})()`

	raw, err := client.Evaluate(ctx, expr, false)
	if err != nil {
		return nil, err
	}
	// raw is a JSON string containing the stringified result.
	var jsonStr string
	if err := json.Unmarshal(raw, &jsonStr); err != nil {
		return nil, fmt.Errorf("parse ui state: %w", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("parse ui state JSON: %w", err)
	}
	return result, nil
}

// Launch starts TradingView with CDP enabled.
func Launch(args LaunchArgs) (map[string]interface{}, error) {
	port := 9222
	if args.Port != nil {
		port = *args.Port
	}
	killExisting := true
	if args.KillExisting != nil {
		killExisting = *args.KillExisting
	}
	tvPath := ""
	if args.TvPath != nil {
		tvPath = *args.TvPath
	}
	return launcher.Launch(port, killExisting, tvPath)
}

// RegisterTools registers all health-group tools into the MCP registry.
func RegisterTools(reg *mcp.Registry) {
	reg.Register(mcp.ToolDef{
		Name:        "tv_health_check",
		Description: "Check CDP connection to TradingView and return current chart state",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			result, err := HealthCheck()
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   err.Error(),
					"hint":    "TradingView is not running with CDP enabled. Use the tv_launch tool to start it automatically.",
				}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "tv_discover",
		Description: "Report which known TradingView API paths are available and their methods",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			result, err := Discover()
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "tv_ui_state",
		Description: "Get current UI state: which panels are open, what buttons are visible/enabled/disabled",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			result, err := UIState()
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "tv_launch",
		Description: "Launch TradingView Desktop with Chrome DevTools Protocol (remote debugging) enabled. Auto-detects install location on Mac, Windows, and Linux.",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"port":          {Type: "number", Description: "CDP port (default 9222)"},
				"kill_existing": {Type: "boolean", Description: "Kill existing TradingView instances first (default true)"},
				"tv_path":       {Type: "string", Description: "Explicit path to TradingView executable (overrides auto-discovery)"},
			},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var a LaunchArgs
			if len(args) > 0 {
				if err := json.Unmarshal(args, &a); err != nil {
					return nil, fmt.Errorf("invalid arguments: %w", err)
				}
			}
			return Launch(a)
		},
	})
}
