// Package pane implements pane_list, pane_set_layout, pane_focus, pane_set_symbol.
package pane

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

const cwc = "window.TradingViewApi._chartWidgetCollection"

// layoutNames mirrors the LAYOUT_NAMES map from pane.js.
var layoutNames = map[string]string{
	"s": "1 chart", "2h": "2 horizontal", "2v": "2 vertical",
	"2-1": "2 top, 1 bottom", "1-2": "1 top, 2 bottom",
	"3h": "3 horizontal", "3v": "3 vertical", "3s": "3 custom",
	"4": "2x2 grid", "4h": "4 horizontal", "4v": "4 vertical", "4s": "4 custom",
	"6": "6 charts", "8": "8 charts", "10": "10 charts",
	"12": "12 charts", "14": "14 charts", "16": "16 charts",
}

// layoutAliases normalises friendly names to layout codes.
var layoutAliases = map[string]string{
	"single": "s", "1": "s", "1x1": "s",
	"2x1": "2h", "1x2": "2v",
	"2x2": "4", "grid": "4", "quad": "4",
	"3x1": "3h", "1x3": "3v",
}

func resolveLayout(layout string) (string, error) {
	code := strings.ToLower(strings.ReplaceAll(layout, " ", ""))
	if a, ok := layoutAliases[code]; ok {
		code = a
	}
	if _, ok := layoutNames[code]; !ok {
		avail := make([]string, 0, len(layoutNames))
		for k, v := range layoutNames {
			avail = append(avail, k+" ("+v+")")
		}
		return "", fmt.Errorf("unknown layout %q; available: %s", layout, strings.Join(avail, ", "))
	}
	return code, nil
}

// ListPanes returns all chart panes with their symbol, resolution, and active state.
func ListPanes() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const expr = `(function() {
		var cwc = ` + cwc + `;
		var layoutType = cwc._layoutType;
		if (typeof layoutType === 'object' && layoutType && typeof layoutType.value === 'function') layoutType = layoutType.value();
		var count = cwc.inlineChartsCount;
		if (typeof count === 'object' && count && typeof count.value === 'function') count = count.value();

		var all = cwc.getAll();
		var panes = [];
		for (var i = 0; i < all.length; i++) {
			try {
				var c = all[i];
				var model = c.model ? c.model() : null;
				var mainSeries = model ? model.mainSeries() : null;
				var sym = mainSeries ? mainSeries.symbol() : 'unknown';
				var res = mainSeries ? mainSeries.interval() : null;
				panes.push({ index: i, symbol: sym, resolution: res || null });
			} catch(e) { panes.push({ index: i, error: e.message }); }
		}

		var activeChart = window.TradingViewApi._activeChartWidgetWV.value();
		var activeIndex = null;
		for (var j = 0; j < all.length; j++) {
			try {
				if (all[j].model && activeChart._chartWidget && all[j] === activeChart._chartWidget) { activeIndex = j; break; }
			} catch(e) {}
		}
		return { layout: layoutType, chart_count: count, active_index: activeIndex, panes: panes };
	})()`

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		raw, err := c.Evaluate(ctx, expr, false)
		if err != nil {
			return err
		}
		var res struct {
			Layout      interface{}   `json:"layout"`
			ChartCount  interface{}   `json:"chart_count"`
			ActiveIndex interface{}   `json:"active_index"`
			Panes       []interface{} `json:"panes"`
		}
		if err := json.Unmarshal(raw, &res); err != nil {
			return fmt.Errorf("parse pane list: %w", err)
		}
		layoutStr := fmt.Sprint(res.Layout)
		result = map[string]interface{}{
			"success":      true,
			"layout":       layoutStr,
			"layout_name":  layoutNames[layoutStr],
			"chart_count":  res.ChartCount,
			"active_index": res.ActiveIndex,
			"panes":        res.Panes,
		}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// SetLayout changes the chart grid layout.
func SetLayout(layout string) (map[string]interface{}, error) {
	code, err := resolveLayout(layout)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	expr := fmt.Sprintf(`%s.setLayout(%s)`, cwc, tv.SafeString(code))

	err = cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		_, err := c.Evaluate(ctx, expr, true)
		return err
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	time.Sleep(500 * time.Millisecond)

	state, _ := ListPanes()
	result := map[string]interface{}{
		"success":     true,
		"layout":      code,
		"layout_name": layoutNames[code],
	}
	if state != nil {
		result["chart_count"] = state["chart_count"]
		result["panes"] = state["panes"]
	}
	return result, nil
}

// FocusPane activates a chart pane by index.
func FocusPane(index int) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	expr := fmt.Sprintf(`(function() {
		var cwc = %s;
		var all = cwc.getAll();
		if (%d >= all.length) return { error: 'Pane index %d out of range (have ' + all.length + ' panes)' };
		var chart = all[%d];
		if (chart._mainDiv) chart._mainDiv.click();
		return { focused: %d, total: all.length };
	})()`, cwc, index, index, index, index)

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		raw, err := c.Evaluate(ctx, expr, false)
		if err != nil {
			return err
		}
		var res map[string]interface{}
		if err := json.Unmarshal(raw, &res); err != nil {
			return fmt.Errorf("parse focus result: %w", err)
		}
		if errMsg, ok := res["error"].(string); ok {
			return fmt.Errorf("%s", errMsg)
		}
		result = map[string]interface{}{
			"success":       true,
			"focused_index": res["focused"],
			"total_panes":   res["total"],
		}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// SetPaneSymbol sets the symbol on a pane by index (focuses first, then sets symbol).
func SetPaneSymbol(index int, symbol string) (map[string]interface{}, error) {
	// Focus the pane first.
	if _, err := FocusPane(index); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	time.Sleep(300 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	setExpr := fmt.Sprintf(`(function() {
		var chart = window.TradingViewApi._activeChartWidgetWV.value();
		return new Promise(function(resolve) {
			chart.setSymbol(%s, {});
			setTimeout(resolve, 500);
		});
	})()`, tv.SafeString(symbol))

	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		_, err := c.Evaluate(ctx, setExpr, true)
		return err
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{"success": true, "index": index, "symbol": symbol}, nil
}

// RegisterTools registers pane_list, pane_set_layout, pane_focus, pane_set_symbol.
func RegisterTools(reg *mcp.Registry) {
	reg.Register(mcp.ToolDef{
		Name:        "pane_list",
		Description: "List all chart panes in the current layout with symbol, resolution, and active state",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			return ListPanes()
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "pane_set_layout",
		Description: "Change the chart grid layout. Codes: s (single), 2h, 2v, 2-1, 1-2, 3h, 3v, 3s, 4 (2x2), 4h, 4v, 4s, 6, 8, 10, 12, 14, 16",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"layout": {Type: "string", Description: "Layout code (e.g. s, 2h, 4) or alias (single, 2x2, quad)"},
			},
			Required: []string{"layout"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Layout string `json:"layout"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			result, err := SetLayout(p.Layout)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "pane_focus",
		Description: "Focus a specific chart pane by index (0-based, from pane_list)",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"index": {Type: "number", Description: "Pane index (0-based)"},
			},
			Required: []string{"index"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Index int `json:"index"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return FocusPane(p.Index)
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "pane_set_symbol",
		Description: "Set the symbol on a specific chart pane by index",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"index":  {Type: "number", Description: "Pane index (0-based)"},
				"symbol": {Type: "string", Description: "Symbol to set (e.g. NQ1!, ES1!, AAPL)"},
			},
			Required: []string{"index", "symbol"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Index  int    `json:"index"`
				Symbol string `json:"symbol"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return SetPaneSymbol(p.Index, p.Symbol)
		},
	})
}
