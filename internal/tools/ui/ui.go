// Package ui implements ui_click, ui_open_panel, ui_fullscreen, ui_keyboard,
// ui_type_text, ui_hover, ui_scroll, ui_mouse_click, ui_find_element,
// ui_evaluate, layout_list, layout_switch.
package ui

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

// keyMap mirrors the keyMap in core/ui.js.
var keyMap = map[string]struct {
	Code string
	VK   int
}{
	"Enter":      {"Enter", 13},
	"Escape":     {"Escape", 27},
	"Tab":        {"Tab", 9},
	"Backspace":  {"Backspace", 8},
	"Delete":     {"Delete", 46},
	"ArrowUp":    {"ArrowUp", 38},
	"ArrowDown":  {"ArrowDown", 40},
	"ArrowLeft":  {"ArrowLeft", 37},
	"ArrowRight": {"ArrowRight", 39},
	"Space":      {"Space", 32},
	"Home":       {"Home", 36},
	"End":        {"End", 35},
	"PageUp":     {"PageUp", 33},
	"PageDown":   {"PageDown", 34},
	"F1":         {"F1", 112},
	"F2":         {"F2", 113},
	"F5":         {"F5", 116},
}

func evalJSON(ctx context.Context, c *cdp.Client, expr string) (interface{}, error) {
	raw, err := c.Evaluate(ctx, expr, false)
	if err != nil {
		return nil, err
	}
	var v interface{}
	_ = json.Unmarshal(raw, &v)
	return v, nil
}

func evalMap(ctx context.Context, c *cdp.Client, expr string) (map[string]interface{}, error) {
	v, err := evalJSON(ctx, c, expr)
	if err != nil {
		return nil, err
	}
	if m, ok := v.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, fmt.Errorf("expected object, got %T", v)
}

// Click clicks a UI element selected by aria-label, data-name, text, or class-contains.
func Click(by, value string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	expr := fmt.Sprintf(`(function() {
		var by = %s;
		var value = %s;
		var el = null;
		if (by === 'aria-label') el = document.querySelector('[aria-label="' + value.replace(/"/g, '\\"') + '"]');
		else if (by === 'data-name') el = document.querySelector('[data-name="' + value.replace(/"/g, '\\"') + '"]');
		else if (by === 'text') {
			var candidates = document.querySelectorAll('button, a, [role="button"], [role="menuitem"], [role="tab"]');
			for (var i = 0; i < candidates.length; i++) {
				var text = candidates[i].textContent.trim();
				if (text === value || text.toLowerCase() === value.toLowerCase()) { el = candidates[i]; break; }
			}
		} else if (by === 'class-contains') el = document.querySelector('[class*="' + value.replace(/"/g, '\\"') + '"]');
		if (!el) return { found: false };
		el.click();
		return { found: true, tag: el.tagName.toLowerCase(), text: (el.textContent || '').trim().substring(0, 80), aria_label: el.getAttribute('aria-label') || null, data_name: el.getAttribute('data-name') || null };
	})()`, tv.SafeString(by), tv.SafeString(value))

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		m, err := evalMap(ctx, c, expr)
		if err != nil {
			return err
		}
		if found, _ := m["found"].(bool); !found {
			return fmt.Errorf("no matching element found for %s=%q", by, value)
		}
		result = map[string]interface{}{"success": true, "clicked": m}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// OpenPanel opens, closes, or toggles a TradingView panel.
func OpenPanel(panel, action string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		var expr string
		if panel == "pine-editor" || panel == "strategy-tester" {
			widgetName := panel
			if panel == "strategy-tester" {
				widgetName = "backtesting"
			}
			expr = fmt.Sprintf(`(function() {
				var bwb = window.TradingView && window.TradingView.bottomWidgetBar;
				if (!bwb) return { error: 'bottomWidgetBar not available' };
				var panel = %s;
				var widgetName = %s;
				var action = %s;
				var bottomArea = document.querySelector('[class*="layout__area--bottom"]');
				var isOpen = !!(bottomArea && bottomArea.offsetHeight > 50);
				if (panel === 'pine-editor') { var monacoEl = document.querySelector('.monaco-editor.pine-editor-monaco'); isOpen = isOpen && !!monacoEl; }
				if (panel === 'strategy-tester') { var stratPanel = document.querySelector('[data-name="backtesting"]') || document.querySelector('[class*="strategyReport"]'); isOpen = isOpen && !!(stratPanel && stratPanel.offsetParent); }
				var performed = 'none';
				if (action === 'open' || (action === 'toggle' && !isOpen)) {
					if (panel === 'pine-editor') { if (typeof bwb.activateScriptEditorTab === 'function') bwb.activateScriptEditorTab(); else if (typeof bwb.showWidget === 'function') bwb.showWidget(widgetName); }
					else { if (typeof bwb.showWidget === 'function') bwb.showWidget(widgetName); }
					performed = 'opened';
				} else if (action === 'close' || (action === 'toggle' && isOpen)) {
					if (typeof bwb.hideWidget === 'function') bwb.hideWidget(widgetName);
					performed = 'closed';
				}
				return { was_open: isOpen, performed: performed };
			})()`, tv.SafeString(panel), tv.SafeString(widgetName), tv.SafeString(action))
		} else {
			type sel struct{ DataName, AriaLabel string }
			selectorMap := map[string]sel{
				"watchlist": {"base-watchlist-widget-button", "Watchlist"},
				"alerts":    {"alerts-button", "Alerts"},
				"trading":   {"trading-button", "Trading Panel"},
			}
			s, ok := selectorMap[panel]
			if !ok {
				return fmt.Errorf("unknown panel %q; valid: pine-editor, strategy-tester, watchlist, alerts, trading", panel)
			}
			expr = fmt.Sprintf(`(function() {
				var dataName = %s;
				var ariaLabel = %s;
				var action = %s;
				var btn = document.querySelector('[data-name="' + dataName + '"]') || document.querySelector('[aria-label="' + ariaLabel + '"]');
				if (!btn) return { error: 'Button not found for panel: ' + %s };
				var isActive = btn.getAttribute('aria-pressed') === 'true' || btn.classList.contains('isActive') || btn.classList.toString().indexOf('active') !== -1 || btn.classList.toString().indexOf('Active') !== -1;
				var rightArea = document.querySelector('[class*="layout__area--right"]');
				var sidebarOpen = !!(rightArea && rightArea.offsetWidth > 50);
				var isOpen = isActive && sidebarOpen;
				var performed = 'none';
				if (action === 'open' && !isOpen) { btn.click(); performed = 'opened'; }
				else if (action === 'close' && isOpen) { btn.click(); performed = 'closed'; }
				else if (action === 'toggle') { btn.click(); performed = isOpen ? 'closed' : 'opened'; }
				else { performed = isOpen ? 'already_open' : 'already_closed'; }
				return { was_open: isOpen, performed: performed };
			})()`,
				tv.SafeString(s.DataName), tv.SafeString(s.AriaLabel),
				tv.SafeString(action), tv.SafeString(panel))
		}

		m, err := evalMap(ctx, c, expr)
		if err != nil {
			return err
		}
		if errMsg, ok := m["error"].(string); ok {
			return fmt.Errorf("%s", errMsg)
		}
		result = map[string]interface{}{
			"success":   true,
			"panel":     panel,
			"action":    action,
			"was_open":  m["was_open"],
			"performed": m["performed"],
		}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// Fullscreen toggles fullscreen mode.
func Fullscreen() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const expr = `(function() {
		var btn = document.querySelector('[data-name="header-toolbar-fullscreen"]');
		if (!btn) return { found: false };
		btn.click();
		return { found: true };
	})()`

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		m, err := evalMap(ctx, c, expr)
		if err != nil {
			return err
		}
		if found, _ := m["found"].(bool); !found {
			return fmt.Errorf("fullscreen button not found")
		}
		result = map[string]interface{}{"success": true, "action": "fullscreen_toggled"}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// LayoutList returns saved chart layouts via getSavedCharts.
func LayoutList() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	const expr = `new Promise(function(resolve) {
		try {
			window.TradingViewApi.getSavedCharts(function(charts) {
				if (!charts || !Array.isArray(charts)) { resolve({layouts: [], source: 'internal_api', error: 'getSavedCharts returned no data'}); return; }
				var result = charts.map(function(c) { return { id: c.id || c.chartId || null, name: c.name || c.title || 'Untitled', symbol: c.symbol || null, resolution: c.resolution || null, modified: c.timestamp || c.modified || null }; });
				resolve({layouts: result, source: 'internal_api'});
			});
			setTimeout(function() { resolve({layouts: [], source: 'internal_api', error: 'getSavedCharts timed out'}); }, 5000);
		} catch(e) { resolve({layouts: [], source: 'internal_api', error: e.message}); }
	})`

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		raw, err := c.Evaluate(ctx, expr, true)
		if err != nil {
			return err
		}
		var layouts struct {
			Layouts []interface{} `json:"layouts"`
			Source  string        `json:"source"`
			Error   string        `json:"error"`
		}
		if err := json.Unmarshal(raw, &layouts); err != nil {
			return fmt.Errorf("parse layout list: %w", err)
		}
		count := 0
		if layouts.Layouts != nil {
			count = len(layouts.Layouts)
		}
		result = map[string]interface{}{
			"success":      true,
			"layout_count": count,
			"source":       layouts.Source,
			"layouts":      layouts.Layouts,
		}
		if layouts.Error != "" {
			result["error"] = layouts.Error
		}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// LayoutSwitch switches to a saved chart layout by name or numeric ID.
func LayoutSwitch(name string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	switchExpr := fmt.Sprintf(`new Promise(function(resolve) {
		try {
			var target = %s;
			if (/^\d+$/.test(target)) { window.TradingViewApi.loadChartFromServer(target); resolve({success: true, method: 'loadChartFromServer', id: target, source: 'internal_api'}); return; }
			window.TradingViewApi.getSavedCharts(function(charts) {
				if (!charts || !Array.isArray(charts)) { resolve({success: false, error: 'getSavedCharts returned no data', source: 'internal_api'}); return; }
				var match = null;
				for (var i = 0; i < charts.length; i++) { var cname = charts[i].name || charts[i].title || ''; if (cname === target || cname.toLowerCase() === target.toLowerCase()) { match = charts[i]; break; } }
				if (!match) { for (var j = 0; j < charts.length; j++) { var cn = (charts[j].name || charts[j].title || '').toLowerCase(); if (cn.indexOf(target.toLowerCase()) !== -1) { match = charts[j]; break; } } }
				if (!match) { resolve({success: false, error: 'Layout "' + target + '" not found.', source: 'internal_api'}); return; }
				var chartId = match.id || match.chartId;
				window.TradingViewApi.loadChartFromServer(chartId);
				resolve({success: true, method: 'loadChartFromServer', id: chartId, name: match.name || match.title, source: 'internal_api'});
			});
			setTimeout(function() { resolve({success: false, error: 'getSavedCharts timed out', source: 'internal_api'}); }, 5000);
		} catch(e) { resolve({success: false, error: e.message, source: 'internal_api'}); }
	})`, tv.SafeString(name))

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		raw, err := c.Evaluate(ctx, switchExpr, true)
		if err != nil {
			return err
		}
		var res map[string]interface{}
		if err := json.Unmarshal(raw, &res); err != nil {
			return fmt.Errorf("parse layout switch: %w", err)
		}
		if success, _ := res["success"].(bool); !success {
			errMsg, _ := res["error"].(string)
			if errMsg == "" {
				errMsg = "unknown error switching layout"
			}
			return fmt.Errorf("%s", errMsg)
		}

		time.Sleep(500 * time.Millisecond)

		// Dismiss "unsaved changes" dialog if present.
		const dismissExpr = `(function() {
			var btns = document.querySelectorAll('button');
			for (var i = 0; i < btns.length; i++) {
				var text = btns[i].textContent.trim();
				if (/open anyway|don't save|discard/i.test(text)) { btns[i].click(); return true; }
			}
			return false;
		})()`
		dismissed := false
		if raw2, err2 := c.Evaluate(ctx, dismissExpr, false); err2 == nil {
			_ = json.Unmarshal(raw2, &dismissed)
		}
		if dismissed {
			time.Sleep(1 * time.Second)
		}

		layoutName, _ := res["name"].(string)
		if layoutName == "" {
			layoutName = name
		}
		result = map[string]interface{}{
			"success":                  true,
			"layout":                   layoutName,
			"layout_id":                res["id"],
			"source":                   res["source"],
			"action":                   "switched",
			"unsaved_dialog_dismissed": dismissed,
		}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// Keyboard dispatches a key event with optional modifiers.
// modifiers is a comma-separated list of: ctrl, alt, shift, meta.
func Keyboard(key string, modifiers []string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mod := 0
	for _, m := range modifiers {
		switch strings.ToLower(m) {
		case "alt":
			mod |= 1
		case "ctrl":
			mod |= 2
		case "meta":
			mod |= 4
		case "shift":
			mod |= 8
		}
	}

	km, ok := keyMap[key]
	if !ok {
		upper := strings.ToUpper(key)
		vk := 0
		if len(upper) == 1 {
			vk = int(upper[0])
		}
		km = struct {
			Code string
			VK   int
		}{"Key" + upper, vk}
	}

	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		if err := c.DispatchKeyEvent(ctx, cdp.KeyEventParams{
			Type:                  "keyDown",
			Key:                   key,
			Code:                  km.Code,
			Modifiers:             mod,
			WindowsVirtualKeyCode: km.VK,
		}); err != nil {
			return err
		}
		return c.DispatchKeyEvent(ctx, cdp.KeyEventParams{
			Type: "keyUp",
			Key:  key,
			Code: km.Code,
		})
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{"success": true, "key": key, "modifiers": modifiers}, nil
}

// TypeText inserts text at the current focus point.
func TypeText(text string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		return c.InsertText(ctx, text)
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	typed := text
	if len(typed) > 100 {
		typed = typed[:100]
	}
	return map[string]interface{}{"success": true, "typed": typed, "length": len(text)}, nil
}

// Hover moves the mouse over a UI element.
func Hover(by, value string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	expr := fmt.Sprintf(`(function() {
		var by = %s;
		var value = %s;
		var el = null;
		if (by === 'aria-label') {
			el = document.querySelector('[aria-label="' + value.replace(/"/g, '\\"') + '"]');
			if (!el) el = document.querySelector('[aria-label*="' + value.replace(/"/g, '\\"') + '"]');
		} else if (by === 'data-name') el = document.querySelector('[data-name="' + value.replace(/"/g, '\\"') + '"]');
		else if (by === 'text') {
			var candidates = document.querySelectorAll('button, a, [role="button"], [role="menuitem"], [role="tab"], span, div');
			for (var i = 0; i < candidates.length; i++) { var text = candidates[i].textContent.trim(); if (text === value || text.toLowerCase() === value.toLowerCase()) { el = candidates[i]; break; } }
		} else if (by === 'class-contains') el = document.querySelector('[class*="' + value.replace(/"/g, '\\"') + '"]');
		if (!el) return null;
		var rect = el.getBoundingClientRect();
		return { x: rect.x + rect.width / 2, y: rect.y + rect.height / 2, tag: el.tagName.toLowerCase() };
	})()`, tv.SafeString(by), tv.SafeString(value))

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		raw, err := c.Evaluate(ctx, expr, false)
		if err != nil {
			return err
		}
		var coords struct {
			X   float64 `json:"x"`
			Y   float64 `json:"y"`
			Tag string  `json:"tag"`
		}
		if err := json.Unmarshal(raw, &coords); err != nil || (coords.X == 0 && coords.Y == 0) {
			// null returned — element not found
			return fmt.Errorf("element not found for %s=%q", by, value)
		}
		if err := c.DispatchMouseEvent(ctx, cdp.MouseEventParams{
			Type: "mouseMoved",
			X:    coords.X,
			Y:    coords.Y,
		}); err != nil {
			return err
		}
		result = map[string]interface{}{
			"success": true,
			"hovered": map[string]interface{}{
				"by":    by,
				"value": value,
				"tag":   coords.Tag,
				"x":     coords.X,
				"y":     coords.Y,
			},
		}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// Scroll scrolls the chart in the given direction by amount pixels (default 300).
func Scroll(direction string, amount int) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if amount <= 0 {
		amount = 300
	}

	const centerExpr = `(function() {
		var el = document.querySelector('[data-name="pane-canvas"]') || document.querySelector('[class*="chart-container"]') || document.querySelector('canvas');
		if (!el) return { x: window.innerWidth / 2, y: window.innerHeight / 2 };
		var rect = el.getBoundingClientRect();
		return { x: rect.x + rect.width / 2, y: rect.y + rect.height / 2 };
	})()`

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		raw, err := c.Evaluate(ctx, centerExpr, false)
		if err != nil {
			return err
		}
		var center struct {
			X float64 `json:"x"`
			Y float64 `json:"y"`
		}
		_ = json.Unmarshal(raw, &center)

		var deltaX, deltaY float64
		px := float64(amount)
		switch direction {
		case "up":
			deltaY = -px
		case "down":
			deltaY = px
		case "left":
			deltaX = -px
		case "right":
			deltaX = px
		default:
			return fmt.Errorf("invalid direction %q; use up, down, left, right", direction)
		}

		if err := c.DispatchMouseEvent(ctx, cdp.MouseEventParams{
			Type:   "mouseWheel",
			X:      center.X,
			Y:      center.Y,
			DeltaX: deltaX,
			DeltaY: deltaY,
		}); err != nil {
			return err
		}
		result = map[string]interface{}{"success": true, "direction": direction, "amount": amount}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// MouseClick clicks at explicit x,y coordinates.
func MouseClick(x, y float64, button string, doubleClick bool) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	btn := button
	if btn != "right" && btn != "middle" {
		btn = "left"
	}
	btnNum := 0
	if btn == "middle" {
		btnNum = 1
	} else if btn == "right" {
		btnNum = 2
	}

	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		if err := c.DispatchMouseEvent(ctx, cdp.MouseEventParams{Type: "mouseMoved", X: x, Y: y}); err != nil {
			return err
		}
		if err := c.DispatchMouseEvent(ctx, cdp.MouseEventParams{
			Type:       "mousePressed",
			X:          x,
			Y:          y,
			Button:     btn,
			Buttons:    btnNum,
			ClickCount: 1,
		}); err != nil {
			return err
		}
		if err := c.DispatchMouseEvent(ctx, cdp.MouseEventParams{Type: "mouseReleased", X: x, Y: y, Button: btn}); err != nil {
			return err
		}
		if doubleClick {
			time.Sleep(50 * time.Millisecond)
			if err := c.DispatchMouseEvent(ctx, cdp.MouseEventParams{
				Type:       "mousePressed",
				X:          x,
				Y:          y,
				Button:     btn,
				Buttons:    btnNum,
				ClickCount: 2,
			}); err != nil {
				return err
			}
			if err := c.DispatchMouseEvent(ctx, cdp.MouseEventParams{Type: "mouseReleased", X: x, Y: y, Button: btn}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{"success": true, "x": x, "y": y, "button": btn, "double_click": doubleClick}, nil
}

// FindElement searches for UI elements by text, aria-label, or CSS selector.
func FindElement(query, strategy string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if strategy == "" {
		strategy = "text"
	}

	expr := fmt.Sprintf(`(function() {
		var query = %s;
		var strategy = %s;
		var results = [];
		if (strategy === 'css') {
			var els = document.querySelectorAll(query);
			for (var i = 0; i < Math.min(els.length, 20); i++) {
				var rect = els[i].getBoundingClientRect();
				results.push({ tag: els[i].tagName.toLowerCase(), text: (els[i].textContent || '').trim().substring(0, 80), aria_label: els[i].getAttribute('aria-label') || null, data_name: els[i].getAttribute('data-name') || null, x: rect.x, y: rect.y, width: rect.width, height: rect.height, visible: els[i].offsetParent !== null });
			}
		} else if (strategy === 'aria-label') {
			var els = document.querySelectorAll('[aria-label*="' + query.replace(/"/g, '\\"') + '"]');
			for (var i = 0; i < Math.min(els.length, 20); i++) {
				var rect = els[i].getBoundingClientRect();
				results.push({ tag: els[i].tagName.toLowerCase(), text: (els[i].textContent || '').trim().substring(0, 80), aria_label: els[i].getAttribute('aria-label') || null, data_name: els[i].getAttribute('data-name') || null, x: rect.x, y: rect.y, width: rect.width, height: rect.height, visible: els[i].offsetParent !== null });
			}
		} else {
			var all = document.querySelectorAll('button, a, [role="button"], [role="menuitem"], [role="tab"], input, select, label, span, div, h1, h2, h3, h4');
			for (var i = 0; i < all.length; i++) {
				var text = all[i].textContent.trim();
				if (text.toLowerCase().indexOf(query.toLowerCase()) !== -1 && text.length < 200) {
					var rect = all[i].getBoundingClientRect();
					if (rect.width > 0 && rect.height > 0) {
						results.push({ tag: all[i].tagName.toLowerCase(), text: text.substring(0, 80), aria_label: all[i].getAttribute('aria-label') || null, data_name: all[i].getAttribute('data-name') || null, x: rect.x, y: rect.y, width: rect.width, height: rect.height, visible: all[i].offsetParent !== null });
						if (results.length >= 20) break;
					}
				}
			}
		}
		return results;
	})()`, tv.SafeString(query), tv.SafeString(strategy))

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		raw, err := c.Evaluate(ctx, expr, false)
		if err != nil {
			return err
		}
		var elements []interface{}
		_ = json.Unmarshal(raw, &elements)
		count := 0
		if elements != nil {
			count = len(elements)
		}
		result = map[string]interface{}{
			"success":  true,
			"query":    query,
			"strategy": strategy,
			"count":    count,
			"elements": elements,
		}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// Evaluate executes arbitrary JavaScript in the TradingView page context.
func Evaluate(expression string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		v, err := evalJSON(ctx, c, expression)
		if err != nil {
			return err
		}
		result = map[string]interface{}{"success": true, "result": v}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// RegisterTools registers all 12 UI + layout MCP tools.
func RegisterTools(reg *mcp.Registry) {
	reg.Register(mcp.ToolDef{
		Name:        "ui_click",
		Description: "Click a UI element by aria-label, data-name, text content, or class substring",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"by":    {Type: "string", Description: "Selector strategy: aria-label, data-name, text, class-contains"},
				"value": {Type: "string", Description: "Value to match against the chosen selector strategy"},
			},
			Required: []string{"by", "value"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				By    string `json:"by"`
				Value string `json:"value"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return Click(p.By, p.Value)
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "ui_open_panel",
		Description: "Open, close, or toggle TradingView panels (pine-editor, strategy-tester, watchlist, alerts, trading)",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"panel":  {Type: "string", Description: "Panel name: pine-editor, strategy-tester, watchlist, alerts, trading"},
				"action": {Type: "string", Description: "Action: open, close, toggle"},
			},
			Required: []string{"panel", "action"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Panel  string `json:"panel"`
				Action string `json:"action"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return OpenPanel(p.Panel, p.Action)
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "ui_fullscreen",
		Description: "Toggle TradingView fullscreen mode",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			return Fullscreen()
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "layout_list",
		Description: "List saved chart layouts",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			return LayoutList()
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "layout_switch",
		Description: "Switch to a saved chart layout by name or ID",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"name": {Type: "string", Description: "Name or ID of the layout to switch to"},
			},
			Required: []string{"name"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Name string `json:"name"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return LayoutSwitch(p.Name)
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "ui_keyboard",
		Description: "Press keyboard keys or shortcuts (e.g., Enter, Escape, Alt+S, Ctrl+Z)",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"key":       {Type: "string", Description: "Key to press (e.g., Enter, Escape, Tab, a, ArrowUp)"},
				"modifiers": {Type: "array", Description: "Modifier keys to hold: ctrl, alt, shift, meta"},
			},
			Required: []string{"key"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Key       string   `json:"key"`
				Modifiers []string `json:"modifiers"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return Keyboard(p.Key, p.Modifiers)
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "ui_type_text",
		Description: "Type text into the currently focused input/textarea element",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"text": {Type: "string", Description: "Text to type into the focused element"},
			},
			Required: []string{"text"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Text string `json:"text"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return TypeText(p.Text)
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "ui_hover",
		Description: "Hover over a UI element by aria-label, data-name, or text content",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"by":    {Type: "string", Description: "Selector strategy: aria-label, data-name, text, class-contains"},
				"value": {Type: "string", Description: "Value to match"},
			},
			Required: []string{"by", "value"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				By    string `json:"by"`
				Value string `json:"value"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return Hover(p.By, p.Value)
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "ui_scroll",
		Description: "Scroll the chart or page up/down/left/right",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"direction": {Type: "string", Description: "Scroll direction: up, down, left, right"},
				"amount":    {Type: "number", Description: "Scroll amount in pixels (default 300)"},
			},
			Required: []string{"direction"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Direction string `json:"direction"`
				Amount    int    `json:"amount"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return Scroll(p.Direction, p.Amount)
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "ui_mouse_click",
		Description: "Click at specific x,y coordinates on the TradingView window",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"x":            {Type: "number", Description: "X coordinate (pixels from left)"},
				"y":            {Type: "number", Description: "Y coordinate (pixels from top)"},
				"button":       {Type: "string", Description: "Mouse button: left, right, middle (default left)"},
				"double_click": {Type: "boolean", Description: "Double click (default false)"},
			},
			Required: []string{"x", "y"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				X           float64 `json:"x"`
				Y           float64 `json:"y"`
				Button      string  `json:"button"`
				DoubleClick bool    `json:"double_click"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return MouseClick(p.X, p.Y, p.Button, p.DoubleClick)
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "ui_find_element",
		Description: "Find UI elements by text, aria-label, or CSS selector and return their positions",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"query":    {Type: "string", Description: "Text content, aria-label value, or CSS selector to search for"},
				"strategy": {Type: "string", Description: "Search strategy: text, aria-label, css (default: text)"},
			},
			Required: []string{"query"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Query    string `json:"query"`
				Strategy string `json:"strategy"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return FindElement(p.Query, p.Strategy)
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "ui_evaluate",
		Description: "Execute JavaScript code in the TradingView page context for advanced automation",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"expression": {Type: "string", Description: "JavaScript expression to evaluate in the page context. Wrap in IIFE for complex logic."},
			},
			Required: []string{"expression"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Expression string `json:"expression"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return Evaluate(p.Expression)
		},
	})
}
