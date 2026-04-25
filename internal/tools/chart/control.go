package chart

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jhonroun/tradingview-mcp-go/internal/cdp"
	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
	tv "github.com/jhonroun/tradingview-mcp-go/internal/tradingview"
)

var chartTypeMap = map[string]int{
	"bars": 0, "candles": 1, "line": 2, "area": 3,
	"renko": 4, "kagi": 5, "pointandfigure": 6, "linebreak": 7,
	"heikinashi": 8, "hollowcandles": 9,
}

// SetSymbol sets the chart symbol and waits for the chart to reload.
func SetSymbol(symbol string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	setExpr := fmt.Sprintf(`new Promise(function(resolve) {
		%s.setSymbol(%s, function() { setTimeout(resolve, 500); });
	})`, tv.ChartAPI, tv.SafeString(symbol))

	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		if _, err := c.Evaluate(ctx, setExpr, true); err != nil {
			return err
		}
		waitForChartReady(ctx, c, symbol)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"success": true, "symbol": symbol}, nil
}

// SetTimeframe changes the chart resolution (e.g. "60", "D", "W").
func SetTimeframe(timeframe string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	expr := fmt.Sprintf(`(function() {
		%s.setResolution(%s, {}); return true;
	})()`, tv.ChartAPI, tv.SafeString(timeframe))

	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		_, err := c.Evaluate(ctx, expr, false)
		return err
	})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"success": true, "timeframe": timeframe}, nil
}

// SetType changes the chart type by name (Candles, Bars, Line, etc.).
func SetType(chartType string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	key := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(chartType, " ", ""), "_", ""))
	typeNum, ok := chartTypeMap[key]
	if !ok {
		return nil, fmt.Errorf("unknown chart type %q; valid: Bars, Candles, Line, Area, Renko, Kagi, PointAndFigure, LineBreak, HeikinAshi, HollowCandles", chartType)
	}

	expr := fmt.Sprintf(`(function() { %s.setChartType(%d); return true; })()`, tv.ChartAPI, typeNum)

	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		_, err := c.Evaluate(ctx, expr, false)
		return err
	})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"success": true, "chartType": chartType, "typeNum": typeNum}, nil
}

// ManageIndicatorArgs holds parameters for chart_manage_indicator.
type ManageIndicatorArgs struct {
	Action   string        `json:"action"`
	Name     string        `json:"name"`
	EntityID string        `json:"entity_id"`
	Inputs   []interface{} `json:"inputs"`
}

// ManageIndicator adds or removes an indicator on the chart.
func ManageIndicator(args ManageIndicatorArgs) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	switch strings.ToLower(args.Action) {
	case "add":
		return addIndicator(ctx, args.Name, args.Inputs)
	case "remove":
		return removeIndicator(ctx, args.EntityID)
	default:
		return nil, fmt.Errorf("unknown action %q; use add or remove", args.Action)
	}
}

func addIndicator(ctx context.Context, name string, inputs []interface{}) (map[string]interface{}, error) {
	inputsJSON, _ := json.Marshal(inputs)
	if inputs == nil {
		inputsJSON = []byte("[]")
	}

	expr := fmt.Sprintf(`(async function() {
		var chart = %s;
		var before = chart.getAllStudies().map(function(s) { return s.id; });
		chart.createStudy(%s, false, false, %s);
		await new Promise(function(r) { setTimeout(r, 1500); });
		var after = chart.getAllStudies().map(function(s) { return s.id; });
		var newIds = after.filter(function(id) { return before.indexOf(id) === -1; });
		return { entityId: newIds[0] || null, name: %s };
	})()`, tv.ChartAPI, tv.SafeString(name), string(inputsJSON), tv.SafeString(name))

	var raw json.RawMessage
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		var err error
		raw, err = c.Evaluate(ctx, expr, true)
		return err
	})
	if err != nil {
		return nil, err
	}
	var result struct {
		EntityID string `json:"entityId"`
		Name     string `json:"name"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse add result: %w", err)
	}
	return map[string]interface{}{
		"success":  true,
		"action":   "add",
		"entityId": result.EntityID,
		"name":     result.Name,
	}, nil
}

func removeIndicator(ctx context.Context, entityID string) (map[string]interface{}, error) {
	if entityID == "" {
		return nil, fmt.Errorf("entity_id is required for remove action")
	}
	expr := fmt.Sprintf(`(function() {
		%s.removeEntity(%s); return true;
	})()`, tv.ChartAPI, tv.SafeString(entityID))

	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		_, err := c.Evaluate(ctx, expr, false)
		return err
	})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"success": true, "action": "remove", "entityId": entityID}, nil
}

// SetVisibleRange zooms the chart to show bars between from and to (Unix seconds).
func SetVisibleRange(from, to int64) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	expr := fmt.Sprintf(`(function() {
		var chart = %s;
		var bars = %s;
		var size = bars.size();
		var fromTs = %d * 1000, toTs = %d * 1000;
		var fromIdx = 0, toIdx = size - 1;
		for (var i = 0; i < size; i++) {
			var bar = bars.get(i);
			if (!bar) continue;
			var t = bar.time;
			if (t <= fromTs) fromIdx = i;
			if (t <= toTs) toIdx = i;
		}
		chart.zoomToBarsRange(fromIdx, toIdx);
		return { success: true, fromIdx: fromIdx, toIdx: toIdx };
	})()`, tv.ChartAPI, tv.BarsPath, from, to)

	var raw json.RawMessage
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		time.Sleep(500 * time.Millisecond)
		var err error
		raw, err = c.Evaluate(ctx, expr, false)
		return err
	})
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse range result: %w", err)
	}
	result["from"] = from
	result["to"] = to
	return result, nil
}

// ScrollToDate scrolls the chart to center on the bar at the given Unix timestamp.
func ScrollToDate(ts int64) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	expr := fmt.Sprintf(`(function() {
		var chart = %s;
		var bars = %s;
		var size = bars.size();
		var target = %d * 1000;
		var idx = 0;
		for (var i = 0; i < size; i++) {
			var bar = bars.get(i);
			if (!bar) continue;
			if (bar.time <= target) idx = i;
		}
		var from = Math.max(0, idx - 25);
		var to = Math.min(size - 1, idx + 25);
		chart.zoomToBarsRange(from, to);
		return { success: true, barIdx: idx };
	})()`, tv.ChartAPI, tv.BarsPath, ts)

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
		return nil, fmt.Errorf("parse scroll result: %w", err)
	}
	result["timestamp"] = ts
	return result, nil
}

func registerControlTools(reg *mcp.Registry) {
	reg.Register(mcp.ToolDef{
		Name:        "chart_set_symbol",
		Description: "Change the chart symbol and wait for it to reload with new data",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"symbol": {Type: "string", Description: "Symbol to set (e.g. BTCUSDT, NASDAQ:AAPL)"},
			},
			Required: []string{"symbol"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Symbol string `json:"symbol"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			result, err := SetSymbol(p.Symbol)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "chart_set_timeframe",
		Description: "Change the chart resolution/timeframe (e.g. 1, 5, 15, 60, 240, D, W, M)",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"timeframe": {Type: "string", Description: "Resolution string: 1, 5, 15, 30, 60, 240, D, W, M"},
			},
			Required: []string{"timeframe"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Timeframe string `json:"timeframe"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			result, err := SetTimeframe(p.Timeframe)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "chart_set_type",
		Description: "Change the chart type: Bars, Candles, Line, Area, Renko, Kagi, PointAndFigure, LineBreak, HeikinAshi, HollowCandles",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"chart_type": {Type: "string", Description: "Chart type name"},
			},
			Required: []string{"chart_type"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				ChartType string `json:"chart_type"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			result, err := SetType(p.ChartType)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "chart_manage_indicator",
		Description: "Add or remove an indicator. action=add requires name; action=remove requires entity_id",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"action":    {Type: "string", Description: "add or remove"},
				"name":      {Type: "string", Description: "Indicator name for add (e.g. RSI, MACD, Bollinger Bands)"},
				"entity_id": {Type: "string", Description: "Entity ID for remove (from chart_get_state)"},
				"inputs":    {Type: "array", Description: "Optional input values array for add"},
			},
			Required: []string{"action"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p ManageIndicatorArgs
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			result, err := ManageIndicator(p)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "chart_set_visible_range",
		Description: "Zoom the chart to show bars between from and to timestamps (Unix seconds)",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"from": {Type: "number", Description: "Start timestamp (Unix seconds)"},
				"to":   {Type: "number", Description: "End timestamp (Unix seconds)"},
			},
			Required: []string{"from", "to"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				From int64 `json:"from"`
				To   int64 `json:"to"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			result, err := SetVisibleRange(p.From, p.To)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "chart_scroll_to_date",
		Description: "Scroll the chart to center on a specific date/timestamp (Unix seconds)",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"timestamp": {Type: "number", Description: "Unix timestamp in seconds"},
			},
			Required: []string{"timestamp"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Timestamp int64 `json:"timestamp"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			result, err := ScrollToDate(p.Timestamp)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})
}
