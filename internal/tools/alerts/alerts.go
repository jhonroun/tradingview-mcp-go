// Package alerts implements alert_create, alert_list, alert_delete,
// watchlist_get, and watchlist_add (P9 group).
package alerts

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jhonroun/tradingview-mcp-go/internal/cdp"
	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
)

// CreateAlert opens the TradingView alert dialog and creates a price alert.
func CreateAlert(condition string, price float64, message string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	priceStr, _ := json.Marshal(fmt.Sprintf("%g", price))
	msgJSON, _ := json.Marshal(message)

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		// Open the alert dialog.
		raw, _ := c.Evaluate(ctx, `(function() {
			var btn = document.querySelector('[aria-label="Create Alert"]')
				|| document.querySelector('[data-name="alerts"]');
			if (btn) { btn.click(); return true; }
			return false;
		})()`, false)

		var opened bool
		if raw == nil || json.Unmarshal(raw, &opened) != nil || !opened {
			// Fallback: Shift+A
			c.DispatchKeyEvent(ctx, cdp.KeyEventParams{Type: "keyDown", Modifiers: 1, Key: "a", Code: "KeyA", WindowsVirtualKeyCode: 65})
			c.DispatchKeyEvent(ctx, cdp.KeyEventParams{Type: "keyUp", Key: "a", Code: "KeyA"})
		}
		time.Sleep(1 * time.Second)

		// Set the price value via React synthetic event override.
		priceSetRaw, _ := c.Evaluate(ctx, `(function() {
			var inputs = document.querySelectorAll('[class*="alert"] input[type="text"], [class*="alert"] input[type="number"]');
			for (var i = 0; i < inputs.length; i++) {
				var label = inputs[i].closest('[class*="row"]') && inputs[i].closest('[class*="row"]').querySelector('[class*="label"]');
				if (label && /value|price/i.test(label.textContent)) {
					var nativeSet = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value').set;
					nativeSet.call(inputs[i], `+string(priceStr)+`);
					inputs[i].dispatchEvent(new Event('input', { bubbles: true }));
					inputs[i].dispatchEvent(new Event('change', { bubbles: true }));
					return true;
				}
			}
			if (inputs.length > 0) {
				var nativeSet = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value').set;
				nativeSet.call(inputs[0], `+string(priceStr)+`);
				inputs[0].dispatchEvent(new Event('input', { bubbles: true }));
				return true;
			}
			return false;
		})()`, false)
		var priceSet bool
		if priceSetRaw != nil {
			json.Unmarshal(priceSetRaw, &priceSet)
		}

		// Set optional message.
		if message != "" {
			c.Evaluate(ctx, `(function() {
				var textarea = document.querySelector('[class*="alert"] textarea')
					|| document.querySelector('textarea[placeholder*="message"]');
				if (textarea) {
					var nativeSet = Object.getOwnPropertyDescriptor(HTMLTextAreaElement.prototype, 'value').set;
					nativeSet.call(textarea, `+string(msgJSON)+`);
					textarea.dispatchEvent(new Event('input', { bubbles: true }));
				}
			})()`, false)
		}

		time.Sleep(500 * time.Millisecond)

		// Click the Create button.
		createdRaw, _ := c.Evaluate(ctx, `(function() {
			var btns = document.querySelectorAll('button[data-name="submit"], button');
			for (var i = 0; i < btns.length; i++) {
				if (/^create$/i.test(btns[i].textContent.trim())) { btns[i].click(); return true; }
			}
			return false;
		})()`, false)
		var created bool
		if createdRaw != nil {
			json.Unmarshal(createdRaw, &created)
		}

		result = map[string]interface{}{
			"success":   created,
			"price":     price,
			"condition": condition,
			"message":   message,
			"price_set": priceSet,
			"source":    "dom_fallback",
		}
		if message == "" {
			result["message"] = "(none)"
		}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// ListAlerts fetches active alerts from TradingView's pricealerts API via CDP.
func ListAlerts() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	const expr = `fetch('https://pricealerts.tradingview.com/list_alerts', { credentials: 'include' })
		.then(function(r) { return r.json(); })
		.then(function(data) {
			if (data.s !== 'ok' || !Array.isArray(data.r)) return { alerts: [], error: data.errmsg || 'Unexpected response' };
			return {
				alerts: data.r.map(function(a) {
					var sym = '';
					try { sym = JSON.parse(a.symbol.replace(/^=/, '')).symbol || a.symbol; } catch(e) { sym = a.symbol; }
					return {
						alert_id: a.alert_id,
						symbol: sym,
						type: a.type,
						message: a.message,
						active: a.active,
						condition: a.condition,
						resolution: a.resolution,
						created: a.create_time,
						last_fired: a.last_fire_time,
						expiration: a.expiration,
					};
				})
			};
		})
		.catch(function(e) { return { alerts: [], error: e.message }; })`

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		raw, err := c.Evaluate(ctx, expr, true)
		if err != nil {
			return err
		}
		var res struct {
			Alerts []interface{} `json:"alerts"`
			Error  string        `json:"error"`
		}
		if err := json.Unmarshal(raw, &res); err != nil {
			return fmt.Errorf("parse alerts list: %w", err)
		}
		result = map[string]interface{}{
			"success":     true,
			"alert_count": len(res.Alerts),
			"source":      "internal_api",
			"alerts":      res.Alerts,
		}
		if res.Error != "" {
			result["error"] = res.Error
		}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// DeleteAlerts opens the alerts context menu for bulk deletion (requires user confirmation).
func DeleteAlerts(deleteAll bool) (map[string]interface{}, error) {
	if !deleteAll {
		return nil, fmt.Errorf("individual alert deletion not yet supported — use delete_all: true")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		raw, _ := c.Evaluate(ctx, `(function() {
			var alertBtn = document.querySelector('[data-name="alerts"]');
			if (alertBtn) alertBtn.click();
			var header = document.querySelector('[data-name="alerts"]');
			if (header) {
				header.dispatchEvent(new MouseEvent('contextmenu', { bubbles: true, clientX: 100, clientY: 100 }));
				return { context_menu_opened: true };
			}
			return { context_menu_opened: false };
		})()`, false)

		var res struct {
			ContextMenuOpened bool `json:"context_menu_opened"`
		}
		if raw != nil {
			json.Unmarshal(raw, &res)
		}
		result = map[string]interface{}{
			"success":             true,
			"note":                "Alert deletion requires manual confirmation in the context menu.",
			"context_menu_opened": res.ContextMenuOpened,
			"source":              "dom_fallback",
		}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// RegisterTools registers alert_create, alert_list, and alert_delete.
func RegisterTools(reg *mcp.Registry) {
	reg.Register(mcp.ToolDef{
		Name:        "alert_create",
		Description: "Create a price alert via the TradingView alert dialog",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"condition": {Type: "string", Description: "Alert condition (e.g. crossing, greater_than, less_than)"},
				"price":     {Type: "number", Description: "Price level for the alert"},
				"message":   {Type: "string", Description: "Optional alert message/note"},
			},
			Required: []string{"condition", "price"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Condition string  `json:"condition"`
				Price     float64 `json:"price"`
				Message   string  `json:"message"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return CreateAlert(p.Condition, p.Price, p.Message)
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "alert_list",
		Description: "List all active price alerts from TradingView",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			return ListAlerts()
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "alert_delete",
		Description: "Delete alerts. Pass delete_all: true to open context menu for bulk deletion (requires confirmation).",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"delete_all": {Type: "boolean", Description: "Set true to delete all alerts via context menu"},
			},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				DeleteAll bool `json:"delete_all"`
			}
			json.Unmarshal(args, &p)
			result, err := DeleteAlerts(p.DeleteAll)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	// Register watchlist tools in same group.
	registerWatchlistTools(reg)
}

func registerWatchlistTools(reg *mcp.Registry) {
	reg.Register(mcp.ToolDef{
		Name:        "watchlist_get",
		Description: "Get all symbols from the current TradingView watchlist with last price, change, and change%",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			return GetWatchlist()
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "watchlist_add",
		Description: "Add a symbol to the TradingView watchlist",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"symbol": {Type: "string", Description: "Symbol to add (e.g. AAPL, BTCUSD, ES1!, NYMEX:CL1!)"},
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
			return AddToWatchlist(p.Symbol)
		},
	})
}

// GetWatchlist reads all symbols from the TradingView watchlist panel.
func GetWatchlist() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const expr = `(function() {
		try {
			var rightArea = document.querySelector('[class*="layout__area--right"]');
			if (!rightArea || rightArea.offsetWidth < 50) return { symbols: [], source: 'panel_closed' };
		} catch(e) {}

		var results = [];
		var seen = {};
		var container = document.querySelector('[class*="layout__area--right"]');
		if (!container) return { symbols: [], source: 'no_container' };

		var symbolEls = container.querySelectorAll('[data-symbol-full]');
		for (var i = 0; i < symbolEls.length; i++) {
			var sym = symbolEls[i].getAttribute('data-symbol-full');
			if (!sym || seen[sym]) continue;
			seen[sym] = true;
			var row = symbolEls[i].closest('[class*="row"]') || symbolEls[i].parentElement;
			var cells = row ? row.querySelectorAll('[class*="cell"], [class*="column"]') : [];
			var nums = [];
			for (var j = 0; j < cells.length; j++) {
				var t = cells[j].textContent.trim();
				if (t && /^[-+]?[\d,]+\.?\d*%?$/.test(t.replace(/[\s,]/g, ''))) nums.push(t);
			}
			results.push({ symbol: sym, last: nums[0] || null, change: nums[1] || null, change_percent: nums[2] || null });
		}
		if (results.length > 0) return { symbols: results, source: 'data_attributes' };

		var items = container.querySelectorAll('[class*="symbolName"], [class*="tickerName"], [class*="symbol-"]');
		for (var k = 0; k < items.length; k++) {
			var text = items[k].textContent.trim();
			if (text && /^[A-Z][A-Z0-9.:!]{0,20}$/.test(text) && !seen[text]) {
				seen[text] = true;
				results.push({ symbol: text, last: null, change: null, change_percent: null });
			}
		}
		return { symbols: results, source: results.length > 0 ? 'text_scan' : 'empty' };
	})()`

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		raw, err := c.Evaluate(ctx, expr, false)
		if err != nil {
			return err
		}
		var res struct {
			Symbols []interface{} `json:"symbols"`
			Source  string        `json:"source"`
		}
		if err := json.Unmarshal(raw, &res); err != nil {
			return fmt.Errorf("parse watchlist: %w", err)
		}
		result = map[string]interface{}{
			"success": true,
			"count":   len(res.Symbols),
			"source":  res.Source,
			"symbols": res.Symbols,
		}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// AddToWatchlist opens the watchlist panel and adds a symbol via text input.
func AddToWatchlist(symbol string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		// Ensure watchlist panel is open.
		panelRaw, _ := c.Evaluate(ctx, `(function() {
			var btn = document.querySelector('[data-name="base-watchlist-widget-button"]')
				|| document.querySelector('[aria-label*="Watchlist"]');
			if (!btn) return { error: 'Watchlist button not found' };
			var isActive = btn.getAttribute('aria-pressed') === 'true'
				|| btn.classList.toString().indexOf('Active') !== -1
				|| btn.classList.toString().indexOf('active') !== -1;
			if (!isActive) { btn.click(); return { opened: true }; }
			return { opened: false };
		})()`, false)

		if panelRaw != nil {
			var panel struct {
				Error  string `json:"error"`
				Opened bool   `json:"opened"`
			}
			if json.Unmarshal(panelRaw, &panel) == nil {
				if panel.Error != "" {
					return fmt.Errorf("%s", panel.Error)
				}
				if panel.Opened {
					time.Sleep(500 * time.Millisecond)
				}
			}
		}

		// Click the Add Symbol button.
		addRaw, _ := c.Evaluate(ctx, `(function() {
			var selectors = [
				'[data-name="add-symbol-button"]',
				'[aria-label="Add symbol"]',
				'[aria-label*="Add symbol"]',
				'button[class*="addSymbol"]',
			];
			for (var s = 0; s < selectors.length; s++) {
				var btn = document.querySelector(selectors[s]);
				if (btn && btn.offsetParent !== null) { btn.click(); return { found: true }; }
			}
			var container = document.querySelector('[class*="layout__area--right"]');
			if (container) {
				var buttons = container.querySelectorAll('button');
				for (var i = 0; i < buttons.length; i++) {
					var label = buttons[i].getAttribute('aria-label') || '';
					if (/add.*symbol/i.test(label) || buttons[i].textContent.trim() === '+') {
						buttons[i].click();
						return { found: true };
					}
				}
			}
			return { found: false };
		})()`, false)

		var addResult struct {
			Found bool `json:"found"`
		}
		if addRaw != nil {
			json.Unmarshal(addRaw, &addResult)
		}
		if !addResult.Found {
			return fmt.Errorf("add symbol button not found in watchlist panel")
		}
		time.Sleep(300 * time.Millisecond)

		// Type symbol and confirm.
		if err := c.InsertText(ctx, symbol); err != nil {
			return fmt.Errorf("insert text: %w", err)
		}
		time.Sleep(500 * time.Millisecond)

		c.DispatchKeyEvent(ctx, cdp.KeyEventParams{Type: "keyDown", Key: "Enter", Code: "Enter", WindowsVirtualKeyCode: 13})
		c.DispatchKeyEvent(ctx, cdp.KeyEventParams{Type: "keyUp", Key: "Enter", Code: "Enter"})
		time.Sleep(300 * time.Millisecond)

		c.DispatchKeyEvent(ctx, cdp.KeyEventParams{Type: "keyDown", Key: "Escape", Code: "Escape", WindowsVirtualKeyCode: 27})
		c.DispatchKeyEvent(ctx, cdp.KeyEventParams{Type: "keyUp", Key: "Escape", Code: "Escape"})

		result = map[string]interface{}{"success": true, "symbol": symbol, "action": "added"}
		return nil
	})
	if err != nil {
		// Try Escape on error to close any open search input.
		cleanCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		cdp.WithSession(cleanCtx, func(c *cdp.Client, _ *cdp.Target) error {
			c.DispatchKeyEvent(cleanCtx, cdp.KeyEventParams{Type: "keyDown", Key: "Escape", Code: "Escape", WindowsVirtualKeyCode: 27})
			c.DispatchKeyEvent(cleanCtx, cdp.KeyEventParams{Type: "keyUp", Key: "Escape", Code: "Escape"})
			return nil
		})
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

