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

	if entityID == "" {
		return map[string]interface{}{"success": false, "error": "entity_id is required"}, nil
	}
	if len(inputs) == 0 {
		return map[string]interface{}{"success": false, "error": "inputs must not be empty", "entityId": entityID, "entity_id": entityID}, nil
	}

	inputsJSON, _ := json.Marshal(inputs)

	expr := fmt.Sprintf(`(async function() {
		var chart = %s;
		var study = chart.getStudyById(%s);
		if (!study) return { success: false, error: 'study not found' };
		var vals = study.getInputValues();
		var overrides = %s;
		var before = {};
		var known = {};
		var requested = Object.keys(overrides || {});
		var matched = [];
		var missing = [];
		function same(a, b) { return JSON.stringify(a) === JSON.stringify(b); }
		function safeValue(v) {
			if (typeof v === 'string' && v.length > 200) return { type: 'string', length: v.length, omitted: true };
			if (v === undefined) return null;
			return v;
		}
		vals.forEach(function(inp) {
			if (!inp || !inp.id) return;
			known[inp.id] = true;
			before[inp.id] = inp.value;
		});
		requested.forEach(function(id) {
			if (known[id]) matched.push(id);
			else missing.push(id);
		});
		if (matched.length === 0) {
			return {
				success: false,
				error: 'no requested input ids found',
				entityId: %s,
				entity_id: %s,
				missing_input_ids: missing,
				requested_input_count: requested.length,
				changed_count: 0,
				source: 'tradingview_study_inputs'
			};
		}
		vals.forEach(function(inp) {
			if (inp && inp.id && overrides[inp.id] !== undefined) inp.value = overrides[inp.id];
		});
		var ret = study.setInputValues(vals);
		if (ret && typeof ret.then === 'function') await ret;
		await new Promise(function(r) { setTimeout(r, 500); });
		var afterVals = study.getInputValues();
		var after = {};
		afterVals.forEach(function(inp) {
			if (!inp || !inp.id) return;
			after[inp.id] = inp.value;
		});
		var changes = [];
		var changed = [];
		var unchanged = [];
		var failed = [];
		matched.forEach(function(id) {
			var beforeVal = before[id];
			var afterVal = after[id];
			var requestedVal = overrides[id];
			var valueChanged = !same(beforeVal, afterVal);
			var applied = same(afterVal, requestedVal);
			if (valueChanged) changed.push(id);
			else unchanged.push(id);
			if (!applied) failed.push(id);
			changes.push({
				id: id,
				before: safeValue(beforeVal),
				requested: safeValue(requestedVal),
				after: safeValue(afterVal),
				changed: valueChanged,
				applied: applied
			});
		});
		var partial = missing.length > 0 || failed.length > 0;
		var out = {
			success: true,
			entityId: %s,
			entity_id: %s,
			source: 'tradingview_study_inputs',
			requested_input_count: requested.length,
			applied_count: matched.length - failed.length,
			changed_count: changed.length,
			changed_inputs: changed,
			unchanged_inputs: unchanged,
			missing_input_ids: missing,
			failed_input_ids: failed,
			changes: changes
		};
		if (partial) {
			out.partial = true;
			out.warning = 'Some requested inputs were missing or did not apply.';
		}
		return out;
	})()`,
		tv.ChartAPI, tv.SafeString(entityID), string(inputsJSON), tv.SafeString(entityID), tv.SafeString(entityID), tv.SafeString(entityID), tv.SafeString(entityID))

	var raw json.RawMessage
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		var err error
		raw, err = c.Evaluate(ctx, expr, true)
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

	if entityID == "" {
		return map[string]interface{}{"success": false, "error": "entity_id is required"}, nil
	}

	visStr := "true"
	if !visible {
		visStr = "false"
	}
	expr := fmt.Sprintf(`(function() {
		var chart = %s;
		var study = chart.getStudyById(%s);
		if (!study) return { success: false, error: 'study not found' };
		var before = study.isVisible();
		study.setVisible(%s);
		var after = study.isVisible();
		return {
			success: true,
			entityId: %s,
			entity_id: %s,
			before_visible: before,
			visible: after,
			changed: before !== after,
			source: 'tradingview_study_visibility'
		};
	})()`,
		tv.ChartAPI, tv.SafeString(entityID), visStr, tv.SafeString(entityID), tv.SafeString(entityID))

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
