// Package indicators implements indicator_set_inputs and indicator_toggle_visibility.
package indicators

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jhonroun/tradingview-mcp-go/internal/cdp"
	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
	tv "github.com/jhonroun/tradingview-mcp-go/internal/tradingview"
)

// SetInputs overrides input values of a study by entity ID.
// inputs is a map of input ID → new value (matches getInputValues() id field).
func SetInputs(entityID string, inputs map[string]interface{}) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	inputsJSON, _ := json.Marshal(inputs)

	expr := fmt.Sprintf(`(function() {
		var chart = %s;
		var study = chart.getStudyById(%s);
		if (!study) return { success: false, error: 'study not found' };
		var vals = study.getInputValues();
		var overrides = %s;
		vals.forEach(function(inp) {
			if (overrides[inp.id] !== undefined) inp.value = overrides[inp.id];
		});
		study.setInputValues(vals);
		return { success: true, entityId: %s };
	})()`,
		tv.ChartAPI, tv.SafeString(entityID), string(inputsJSON), tv.SafeString(entityID))

	var raw json.RawMessage
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		var err error
		raw, err = c.Evaluate(ctx, expr, false)
		return err
	})
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse set inputs: %w", err)
	}
	return result, nil
}

// ToggleVisibility shows or hides an indicator by entity ID.
func ToggleVisibility(entityID string, visible bool) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	visStr := "true"
	if !visible {
		visStr = "false"
	}
	expr := fmt.Sprintf(`(function() {
		var chart = %s;
		var study = chart.getStudyById(%s);
		if (!study) return { success: false, error: 'study not found' };
		study.setVisible(%s);
		return { success: true, entityId: %s, visible: study.isVisible() };
	})()`,
		tv.ChartAPI, tv.SafeString(entityID), visStr, tv.SafeString(entityID))

	var raw json.RawMessage
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		var err error
		raw, err = c.Evaluate(ctx, expr, false)
		return err
	})
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse toggle: %w", err)
	}
	return result, nil
}

// RegisterTools registers indicator_set_inputs and indicator_toggle_visibility.
func RegisterTools(reg *mcp.Registry) {
	reg.Register(mcp.ToolDef{
		Name:        "indicator_set_inputs",
		Description: "Override input values of an existing indicator. Use chart_get_state to get entity IDs and input IDs.",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"entity_id": {Type: "string", Description: "Study entity ID from chart_get_state"},
				"inputs":    {Type: "object", Description: "Map of input ID to new value"},
			},
			Required: []string{"entity_id", "inputs"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				EntityID string                 `json:"entity_id"`
				Inputs   map[string]interface{} `json:"inputs"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			result, err := SetInputs(p.EntityID, p.Inputs)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "indicator_toggle_visibility",
		Description: "Show or hide an indicator on the chart without removing it.",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"entity_id": {Type: "string", Description: "Study entity ID from chart_get_state"},
				"visible":   {Type: "boolean", Description: "true to show, false to hide"},
			},
			Required: []string{"entity_id", "visible"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				EntityID string `json:"entity_id"`
				Visible  bool   `json:"visible"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			result, err := ToggleVisibility(p.EntityID, p.Visible)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})
}
