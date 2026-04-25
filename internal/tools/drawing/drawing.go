// Package drawing implements draw_shape, draw_list, draw_get_properties,
// draw_remove_one, and draw_clear MCP tools.
package drawing

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/jhonroun/tradingview-mcp-go/internal/cdp"
	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
	tv "github.com/jhonroun/tradingview-mcp-go/internal/tradingview"
)

// DrawPoint is a {time, price} coordinate for a chart drawing.
type DrawPoint struct {
	Time  float64 `json:"time"`
	Price float64 `json:"price"`
}

// DrawShapeArgs holds parameters for draw_shape.
type DrawShapeArgs struct {
	Shape     string                 `json:"shape"`
	Point     DrawPoint              `json:"point"`
	Point2    *DrawPoint             `json:"point2,omitempty"`
	Overrides map[string]interface{} `json:"overrides,omitempty"`
	Text      string                 `json:"text,omitempty"`
}

func fmtNum(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func requireFinite(v float64, name string) error {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return fmt.Errorf("%s must be a finite number", name)
	}
	return nil
}

const getAllShapesExpr = tv.ChartAPI + `.getAllShapes().map(function(s) { return s.id; })`

// DrawShape creates a shape on the chart and returns the new entity ID.
func DrawShape(args DrawShapeArgs) (map[string]interface{}, error) {
	if err := requireFinite(args.Point.Time, "point.time"); err != nil {
		return nil, err
	}
	if err := requireFinite(args.Point.Price, "point.price"); err != nil {
		return nil, err
	}
	if args.Point2 != nil {
		if err := requireFinite(args.Point2.Time, "point2.time"); err != nil {
			return nil, err
		}
		if err := requireFinite(args.Point2.Price, "point2.price"); err != nil {
			return nil, err
		}
	}

	overridesJSON, _ := json.Marshal(args.Overrides)
	if args.Overrides == nil {
		overridesJSON = []byte("{}")
	}
	textJSON, _ := json.Marshal(args.Text)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		beforeRaw, err := c.Evaluate(ctx, getAllShapesExpr, false)
		if err != nil {
			return err
		}
		var before []string
		json.Unmarshal(beforeRaw, &before)

		var createExpr string
		if args.Point2 != nil {
			createExpr = fmt.Sprintf(
				`%s.createMultipointShape([{time:%s,price:%s},{time:%s,price:%s}],{shape:%s,overrides:%s,text:%s})`,
				tv.ChartAPI,
				fmtNum(args.Point.Time), fmtNum(args.Point.Price),
				fmtNum(args.Point2.Time), fmtNum(args.Point2.Price),
				tv.SafeString(args.Shape), string(overridesJSON), string(textJSON),
			)
		} else {
			createExpr = fmt.Sprintf(
				`%s.createShape({time:%s,price:%s},{shape:%s,overrides:%s,text:%s})`,
				tv.ChartAPI,
				fmtNum(args.Point.Time), fmtNum(args.Point.Price),
				tv.SafeString(args.Shape), string(overridesJSON), string(textJSON),
			)
		}
		c.Evaluate(ctx, createExpr, false)
		time.Sleep(200 * time.Millisecond)

		afterRaw, err := c.Evaluate(ctx, getAllShapesExpr, false)
		if err != nil {
			return err
		}
		var after []string
		json.Unmarshal(afterRaw, &after)

		beforeSet := make(map[string]bool, len(before))
		for _, id := range before {
			beforeSet[id] = true
		}
		newID := ""
		for _, id := range after {
			if !beforeSet[id] {
				newID = id
				break
			}
		}
		result = map[string]interface{}{
			"success":   true,
			"shape":     args.Shape,
			"entity_id": newID,
		}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// ListDrawings returns all shapes currently on the chart.
func ListDrawings() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const expr = `(function() {
		var api = ` + tv.ChartAPI + `;
		var all = api.getAllShapes();
		return all.map(function(s) { return { id: s.id, name: s.name }; });
	})()`

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		raw, err := c.Evaluate(ctx, expr, false)
		if err != nil {
			return err
		}
		var shapes []interface{}
		json.Unmarshal(raw, &shapes)
		result = map[string]interface{}{
			"success": true,
			"count":   len(shapes),
			"shapes":  shapes,
		}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// GetProperties returns all available properties of a shape by entity ID.
func GetProperties(entityID string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	expr := fmt.Sprintf(`(function() {
		var api = %s;
		var eid = %s;
		var props = { entity_id: eid };
		var shape = api.getShapeById(eid);
		if (!shape) return { error: 'Shape not found: ' + eid };
		var methods = [];
		try { for (var key in shape) { if (typeof shape[key] === 'function') methods.push(key); } props.available_methods = methods; } catch(e) {}
		try { var pts = shape.getPoints(); if (pts) props.points = pts; } catch(e) { props.points_error = e.message; }
		try { var ovr = shape.getProperties(); if (ovr) props.properties = ovr; } catch(e) {
			try { var ovr2 = shape.properties(); if (ovr2) props.properties = ovr2; } catch(e2) { props.properties_error = e2.message; }
		}
		try { props.visible = shape.isVisible(); } catch(e) {}
		try { props.locked = shape.isLocked(); } catch(e) {}
		try { props.selectable = shape.isSelectionEnabled(); } catch(e) {}
		try {
			var all = api.getAllShapes();
			for (var i = 0; i < all.length; i++) { if (all[i].id === eid) { props.name = all[i].name; break; } }
		} catch(e) {}
		return props;
	})()`, tv.ChartAPI, tv.SafeString(entityID))

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		raw, err := c.Evaluate(ctx, expr, false)
		if err != nil {
			return err
		}
		var res map[string]interface{}
		if err := json.Unmarshal(raw, &res); err != nil {
			return fmt.Errorf("parse properties: %w", err)
		}
		if errMsg, ok := res["error"].(string); ok {
			return fmt.Errorf("%s", errMsg)
		}
		res["success"] = true
		result = res
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// RemoveOne removes a single shape by entity ID.
func RemoveOne(entityID string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	expr := fmt.Sprintf(`(function() {
		var api = %s;
		var eid = %s;
		var before = api.getAllShapes();
		var found = false;
		for (var i = 0; i < before.length; i++) { if (before[i].id === eid) { found = true; break; } }
		if (!found) return { removed: false, error: 'Shape not found: ' + eid, available: before.map(function(s) { return s.id; }) };
		api.removeEntity(eid);
		var after = api.getAllShapes();
		var stillExists = false;
		for (var j = 0; j < after.length; j++) { if (after[j].id === eid) { stillExists = true; break; } }
		return { removed: !stillExists, entity_id: eid, remaining_shapes: after.length };
	})()`, tv.ChartAPI, tv.SafeString(entityID))

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		raw, err := c.Evaluate(ctx, expr, false)
		if err != nil {
			return err
		}
		var res map[string]interface{}
		if err := json.Unmarshal(raw, &res); err != nil {
			return fmt.Errorf("parse remove result: %w", err)
		}
		if errMsg, ok := res["error"].(string); ok {
			return fmt.Errorf("%s", errMsg)
		}
		res["success"] = true
		result = res
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// ClearAll removes all shapes from the chart.
func ClearAll() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		_, err := c.Evaluate(ctx, tv.ChartAPI+`.removeAllShapes()`, false)
		return err
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{"success": true, "action": "all_shapes_removed"}, nil
}

// RegisterTools registers draw_shape, draw_list, draw_get_properties,
// draw_remove_one, and draw_clear into the MCP registry.
func RegisterTools(reg *mcp.Registry) {
	reg.Register(mcp.ToolDef{
		Name:        "draw_shape",
		Description: "Draw a shape on the chart. shape: horizontal_line, trend_line, rectangle, text, arrow_up, arrow_down, flag, circle, triangle, etc.",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"shape":     {Type: "string", Description: "Shape type name (e.g. horizontal_line, trend_line, rectangle, text)"},
				"point":     {Type: "object", Description: "{time: unix_seconds, price: float}"},
				"point2":    {Type: "object", Description: "Second point for multi-point shapes like trend_line or rectangle"},
				"overrides": {Type: "object", Description: "Style overrides map (e.g. {linecolor: '#FF0000', linewidth: 2})"},
				"text":      {Type: "string", Description: "Text label for text/label shapes"},
			},
			Required: []string{"shape", "point"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p DrawShapeArgs
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return DrawShape(p)
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "draw_list",
		Description: "List all drawing shapes currently on the chart with their entity IDs and names",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			return ListDrawings()
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "draw_get_properties",
		Description: "Get all properties of a drawing shape by entity ID: points, style overrides, visibility, lock state",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"entity_id": {Type: "string", Description: "Drawing entity ID from draw_list"},
			},
			Required: []string{"entity_id"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				EntityID string `json:"entity_id"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return GetProperties(p.EntityID)
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "draw_remove_one",
		Description: "Remove a single drawing shape by entity ID",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"entity_id": {Type: "string", Description: "Drawing entity ID from draw_list"},
			},
			Required: []string{"entity_id"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				EntityID string `json:"entity_id"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return RemoveOne(p.EntityID)
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "draw_clear",
		Description: "Remove all drawing shapes from the chart",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			return ClearAll()
		},
	})
}
