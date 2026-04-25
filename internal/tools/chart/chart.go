// Package chart implements chart_get_state and chart_get_visible_range.
package chart

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jhonroun/tradingview-mcp-go/internal/cdp"
	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
	tv "github.com/jhonroun/tradingview-mcp-go/internal/tradingview"
)

// StudyInfo mirrors the {id, name} shape returned by getAllStudies().
type StudyInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GetState returns the current chart symbol, resolution, type, and all studies.
func GetState() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const expr = `(function() {
		var chart = ` + tv.ChartAPI + `;
		var studies = [];
		try {
			var allStudies = chart.getAllStudies();
			studies = allStudies.map(function(s) {
				return { id: s.id, name: s.name || s.title || 'unknown' };
			});
		} catch(e) {}
		return {
			symbol: chart.symbol(),
			resolution: chart.resolution(),
			chartType: chart.chartType(),
			studies: studies,
		};
	})()`

	var raw json.RawMessage
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		var err error
		raw, err = c.Evaluate(ctx, expr, false)
		return err
	})
	if err != nil {
		return nil, err
	}

	var state struct {
		Symbol     string      `json:"symbol"`
		Resolution string      `json:"resolution"`
		ChartType  interface{} `json:"chartType"`
		Studies    []StudyInfo `json:"studies"`
	}
	if err := json.Unmarshal(raw, &state); err != nil {
		return nil, fmt.Errorf("parse chart state: %w", err)
	}
	return map[string]interface{}{
		"success":    true,
		"symbol":     state.Symbol,
		"resolution": state.Resolution,
		"chartType":  state.ChartType,
		"studies":    state.Studies,
	}, nil
}

// GetVisibleRange returns the chart's visible time range and bars range.
func GetVisibleRange() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const expr = `(function() {
		var chart = ` + tv.ChartAPI + `;
		return {
			visible_range: chart.getVisibleRange(),
			bars_range: chart.getVisibleBarsRange(),
		};
	})()`

	var raw json.RawMessage
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		var err error
		raw, err = c.Evaluate(ctx, expr, false)
		return err
	})
	if err != nil {
		return nil, err
	}

	var result struct {
		VisibleRange interface{} `json:"visible_range"`
		BarsRange    interface{} `json:"bars_range"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse visible range: %w", err)
	}
	return map[string]interface{}{
		"success":       true,
		"visible_range": result.VisibleRange,
		"bars_range":    result.BarsRange,
	}, nil
}

// RegisterTools registers all chart tools (read-only P5 + control P6) into the MCP registry.
func RegisterTools(reg *mcp.Registry) {
	reg.Register(mcp.ToolDef{
		Name:        "chart_get_state",
		Description: "Get current chart state: symbol, timeframe, chart type, and list of all active indicators with entity IDs",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			result, err := GetState()
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "chart_get_visible_range",
		Description: "Get the currently visible time range and bars range of the chart",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			result, err := GetVisibleRange()
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	registerControlTools(reg)
	registerSymbolTools(reg)
}
