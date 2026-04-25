// Package pine implements Pine Script MCP tools: get/set source, compile,
// save, errors, console, new, open, list, check, and smart_compile.
package pine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jhonroun/tradingview-mcp-go/internal/cdp"
	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
)

// findMonaco is the React-Fiber-based Monaco editor finder injected into the TV page.
const findMonaco = `(function findMonacoEditor() {
    var container = document.querySelector('.monaco-editor.pine-editor-monaco');
    if (!container) return null;
    var el = container;
    var fiberKey;
    for (var i = 0; i < 20; i++) {
        if (!el) break;
        fiberKey = Object.keys(el).find(function(k) { return k.startsWith('__reactFiber$'); });
        if (fiberKey) break;
        el = el.parentElement;
    }
    if (!fiberKey) return null;
    var current = el[fiberKey];
    for (var d = 0; d < 15; d++) {
        if (!current) break;
        if (current.memoizedProps && current.memoizedProps.value && current.memoizedProps.value.monacoEnv) {
            var env = current.memoizedProps.value.monacoEnv;
            if (env.editor && typeof env.editor.getEditors === 'function') {
                var editors = env.editor.getEditors();
                if (editors.length > 0) return { editor: editors[0], env: env };
            }
        }
        current = current.return;
    }
    return null;
})()`

// ensurePineEditorOpen opens the Pine Script editor if needed and polls until Monaco is ready.
func ensurePineEditorOpen(ctx context.Context, c *cdp.Client) bool {
	// Check if Monaco is already accessible.
	if monacoReady(ctx, c) {
		return true
	}

	// Try TradingView bottomWidgetBar API.
	c.Evaluate(ctx, `(function() {
		var bwb = window.TradingView && window.TradingView.bottomWidgetBar;
		if (!bwb) return;
		if (typeof bwb.activateScriptEditorTab === 'function') bwb.activateScriptEditorTab();
		else if (typeof bwb.showWidget === 'function') bwb.showWidget('pine-editor');
	})()`, false)

	// Try clicking the Pine button in the DOM.
	c.Evaluate(ctx, `(function() {
		var btn = document.querySelector('[aria-label="Pine"]')
			|| document.querySelector('[data-name="pine-dialog-button"]');
		if (btn) btn.click();
	})()`, false)

	// Poll up to 10 s.
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return false
		default:
		}
		time.Sleep(200 * time.Millisecond)
		if monacoReady(ctx, c) {
			return true
		}
	}
	return false
}

func monacoReady(ctx context.Context, c *cdp.Client) bool {
	raw, err := c.Evaluate(ctx, `(function() { return `+findMonaco+` !== null; })()`, false)
	if err != nil {
		return false
	}
	var ready bool
	return json.Unmarshal(raw, &ready) == nil && ready
}

// ── CDP-dependent tools ───────────────────────────────────────────────────────

// GetSource returns the current Pine Script source code from the Monaco editor.
func GetSource() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		if !ensurePineEditorOpen(ctx, c) {
			return fmt.Errorf("could not open Pine Editor or Monaco not found in React fiber tree")
		}
		raw, err := c.Evaluate(ctx, `(function() {
			var m = `+findMonaco+`;
			if (!m) return null;
			return m.editor.getValue();
		})()`, false)
		if err != nil {
			return err
		}
		var source string
		if json.Unmarshal(raw, &source) != nil {
			return fmt.Errorf("Monaco editor getValue() returned unexpected type")
		}
		result = map[string]interface{}{
			"success":     true,
			"source":      source,
			"line_count":  len(strings.Split(source, "\n")),
			"char_count":  len(source),
		}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// SetSource injects source code into the Monaco editor.
func SetSource(source string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	escaped, _ := json.Marshal(source)

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		if !ensurePineEditorOpen(ctx, c) {
			return fmt.Errorf("could not open Pine Editor")
		}
		raw, err := c.Evaluate(ctx, `(function() {
			var m = `+findMonaco+`;
			if (!m) return false;
			m.editor.setValue(`+string(escaped)+`);
			return true;
		})()`, false)
		if err != nil {
			return err
		}
		var ok bool
		if json.Unmarshal(raw, &ok) != nil || !ok {
			return fmt.Errorf("Monaco found but setValue() failed")
		}
		result = map[string]interface{}{
			"success":   true,
			"lines_set": len(strings.Split(source, "\n")),
		}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

const compileButtonJS = `(function() {
    var btns = document.querySelectorAll('button');
    var fallback = null;
    var saveBtn = null;
    for (var i = 0; i < btns.length; i++) {
        var text = btns[i].textContent.trim();
        if (/save and add to chart/i.test(text)) {
            btns[i].click();
            return 'Save and add to chart';
        }
        if (!fallback && /^(Add to chart|Update on chart)/i.test(text)) {
            fallback = btns[i];
        }
        if (!saveBtn && btns[i].className.indexOf('saveButton') !== -1 && btns[i].offsetParent !== null) {
            saveBtn = btns[i];
        }
    }
    if (fallback) { fallback.click(); return fallback.textContent.trim(); }
    if (saveBtn) { saveBtn.click(); return 'Pine Save'; }
    return null;
})()`

// Compile clicks the compile/add-to-chart button or falls back to Ctrl+Enter.
func Compile() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		if !ensurePineEditorOpen(ctx, c) {
			return fmt.Errorf("could not open Pine Editor")
		}
		raw, _ := c.Evaluate(ctx, compileButtonJS, false)
		var clicked *string
		if raw != nil {
			var v interface{}
			if json.Unmarshal(raw, &v) == nil {
				if s, ok := v.(string); ok {
					clicked = &s
				}
			}
		}
		if clicked == nil {
			c.DispatchKeyEvent(ctx, cdp.KeyEventParams{Type: "keyDown", Modifiers: 2, Key: "Enter", Code: "Enter", WindowsVirtualKeyCode: 13})
			c.DispatchKeyEvent(ctx, cdp.KeyEventParams{Type: "keyUp", Key: "Enter", Code: "Enter"})
		}
		time.Sleep(2 * time.Second)

		btn := "keyboard_shortcut"
		if clicked != nil {
			btn = *clicked
		}
		result = map[string]interface{}{
			"success":        true,
			"button_clicked": btn,
		}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// SmartCompile compiles and reports errors plus whether a new study was added.
func SmartCompile() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	const countStudiesJS = `(function() {
		try {
			var chart = window.TradingViewApi._activeChartWidgetWV.value();
			if (chart && typeof chart.getAllStudies === 'function') return chart.getAllStudies().length;
		} catch(e) {}
		return null;
	})()`

	const smartButtonJS = `(function() {
		var btns = document.querySelectorAll('button');
		var addBtn = null;
		var updateBtn = null;
		var saveBtn = null;
		for (var i = 0; i < btns.length; i++) {
			var text = btns[i].textContent.trim();
			if (/save and add to chart/i.test(text)) { btns[i].click(); return 'Save and add to chart'; }
			if (!addBtn && /^add to chart$/i.test(text)) addBtn = btns[i];
			if (!updateBtn && /^update on chart$/i.test(text)) updateBtn = btns[i];
			if (!saveBtn && btns[i].className.indexOf('saveButton') !== -1 && btns[i].offsetParent !== null) saveBtn = btns[i];
		}
		if (addBtn) { addBtn.click(); return 'Add to chart'; }
		if (updateBtn) { updateBtn.click(); return 'Update on chart'; }
		if (saveBtn) { saveBtn.click(); return 'Pine Save'; }
		return null;
	})()`

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		if !ensurePineEditorOpen(ctx, c) {
			return fmt.Errorf("could not open Pine Editor")
		}
		beforeRaw, _ := c.Evaluate(ctx, countStudiesJS, false)
		var before *int
		if beforeRaw != nil {
			var n int
			if json.Unmarshal(beforeRaw, &n) == nil {
				before = &n
			}
		}

		raw, _ := c.Evaluate(ctx, smartButtonJS, false)
		var clicked *string
		if raw != nil {
			var v interface{}
			if json.Unmarshal(raw, &v) == nil {
				if s, ok := v.(string); ok {
					clicked = &s
				}
			}
		}
		if clicked == nil {
			c.DispatchKeyEvent(ctx, cdp.KeyEventParams{Type: "keyDown", Modifiers: 2, Key: "Enter", Code: "Enter", WindowsVirtualKeyCode: 13})
			c.DispatchKeyEvent(ctx, cdp.KeyEventParams{Type: "keyUp", Key: "Enter", Code: "Enter"})
		}
		time.Sleep(2500 * time.Millisecond)

		errorsRaw, _ := c.Evaluate(ctx, getMarkersJS(), false)
		var errors []interface{}
		if errorsRaw != nil {
			json.Unmarshal(errorsRaw, &errors)
		}

		afterRaw, _ := c.Evaluate(ctx, countStudiesJS, false)
		var studyAdded interface{}
		if before != nil && afterRaw != nil {
			var after int
			if json.Unmarshal(afterRaw, &after) == nil {
				studyAdded = after > *before
			}
		}

		btn := "keyboard_shortcut"
		if clicked != nil {
			btn = *clicked
		}
		result = map[string]interface{}{
			"success":        true,
			"button_clicked": btn,
			"has_errors":     len(errors) > 0,
			"errors":         errors,
			"study_added":    studyAdded,
		}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

func getMarkersJS() string {
	return `(function() {
		var m = ` + findMonaco + `;
		if (!m) return [];
		var model = m.editor.getModel();
		if (!model) return [];
		var markers = m.env.editor.getModelMarkers({ resource: model.uri });
		return markers.map(function(mk) {
			return { line: mk.startLineNumber, column: mk.startColumn, message: mk.message, severity: mk.severity };
		});
	})()`
}

// GetErrors returns Monaco model markers (compilation errors/warnings).
func GetErrors() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		if !ensurePineEditorOpen(ctx, c) {
			return fmt.Errorf("could not open Pine Editor")
		}
		raw, err := c.Evaluate(ctx, getMarkersJS(), false)
		if err != nil {
			return err
		}
		var errors []interface{}
		json.Unmarshal(raw, &errors)
		result = map[string]interface{}{
			"success":     true,
			"has_errors":  len(errors) > 0,
			"error_count": len(errors),
			"errors":      errors,
		}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// Save saves the current Pine Script via Ctrl+S, handling the save dialog if needed.
func Save() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		if !ensurePineEditorOpen(ctx, c) {
			return fmt.Errorf("could not open Pine Editor")
		}
		c.DispatchKeyEvent(ctx, cdp.KeyEventParams{Type: "keyDown", Modifiers: 2, Key: "s", Code: "KeyS", WindowsVirtualKeyCode: 83})
		c.DispatchKeyEvent(ctx, cdp.KeyEventParams{Type: "keyUp", Key: "s", Code: "KeyS"})
		time.Sleep(800 * time.Millisecond)

		raw, _ := c.Evaluate(ctx, `(function() {
			var btns = document.querySelectorAll('button');
			for (var i = 0; i < btns.length; i++) {
				var text = btns[i].textContent.trim();
				if (text === 'Save' && btns[i].offsetParent !== null) {
					var parent = btns[i].closest('[class*="dialog"], [class*="modal"], [class*="popup"], [role="dialog"]');
					if (parent) { btns[i].click(); return true; }
				}
			}
			return false;
		})()`, false)

		action := "Ctrl+S_dispatched"
		var dialogHandled bool
		if raw != nil && json.Unmarshal(raw, &dialogHandled) == nil && dialogHandled {
			action = "saved_with_dialog"
			time.Sleep(500 * time.Millisecond)
		}
		result = map[string]interface{}{"success": true, "action": action}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// GetConsole reads Pine Script console/log output from the DOM.
func GetConsole() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	const consoleJS = `(function() {
		var results = [];
		var rows = document.querySelectorAll('[class*="consoleRow"], [class*="log-"], [class*="consoleLine"]');
		if (rows.length === 0) {
			var bottomArea = document.querySelector('[class*="layout__area--bottom"]')
				|| document.querySelector('[class*="bottom-widgetbar-content"]');
			if (bottomArea) {
				rows = bottomArea.querySelectorAll('[class*="message"], [class*="log"], [class*="console"]');
			}
		}
		if (rows.length === 0) {
			var pinePanel = document.querySelector('.pine-editor-container')
				|| document.querySelector('[class*="pine-editor"]')
				|| document.querySelector('[class*="layout__area--bottom"]');
			if (pinePanel) {
				var allSpans = pinePanel.querySelectorAll('span, div');
				var arr = [];
				for (var s = 0; s < allSpans.length; s++) {
					var txt = allSpans[s].textContent.trim();
					if (/^\d{2}:\d{2}:\d{2}/.test(txt) || /error|warning|info/i.test(allSpans[s].className)) {
						arr.push(allSpans[s]);
					}
				}
				rows = arr;
			}
		}
		for (var i = 0; i < rows.length; i++) {
			var text = rows[i].textContent.trim();
			if (!text) continue;
			var ts = null;
			var tsMatch = text.match(/^(\d{4}-\d{2}-\d{2}\s+)?\d{2}:\d{2}:\d{2}/);
			if (tsMatch) ts = tsMatch[0];
			var type = 'info';
			var cls = rows[i].className || '';
			if (/error/i.test(cls) || /error/i.test(text.substring(0, 30))) type = 'error';
			else if (/compil/i.test(text.substring(0, 40))) type = 'compile';
			else if (/warn/i.test(cls)) type = 'warning';
			results.push({ timestamp: ts, type: type, message: text });
		}
		return results;
	})()`

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		if !ensurePineEditorOpen(ctx, c) {
			return fmt.Errorf("could not open Pine Editor")
		}
		raw, err := c.Evaluate(ctx, consoleJS, false)
		if err != nil {
			return err
		}
		var entries []interface{}
		json.Unmarshal(raw, &entries)
		result = map[string]interface{}{
			"success":     true,
			"entries":     entries,
			"entry_count": len(entries),
		}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// NewScript creates a blank Pine Script of the given type in the editor.
func NewScript(scriptType string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	templates := map[string]string{
		"indicator": "//@version=6\nindicator(\"My script\")\nplot(close)",
		"strategy":  "//@version=6\nstrategy(\"My strategy\", overlay=true)\n",
		"library":   "//@version=6\n// @description TODO: add library description here\nlibrary(\"MyLibrary\")\n",
	}
	t, ok := templates[scriptType]
	if !ok {
		t = templates["indicator"]
		scriptType = "indicator"
	}
	escaped, _ := json.Marshal(t)

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		if !ensurePineEditorOpen(ctx, c) {
			return fmt.Errorf("could not open Pine Editor")
		}
		raw, err := c.Evaluate(ctx, `(function() {
			var m = `+findMonaco+`;
			if (!m) return false;
			m.editor.setValue(`+string(escaped)+`);
			return true;
		})()`, false)
		if err != nil {
			return err
		}
		var ok bool
		if json.Unmarshal(raw, &ok) != nil || !ok {
			return fmt.Errorf("Monaco editor not found")
		}
		result = map[string]interface{}{
			"success":             true,
			"type":                scriptType,
			"action":              "new_script_created",
		}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// OpenScript opens a saved Pine Script by name from the pine-facade API.
func OpenScript(name string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	escapedName, _ := json.Marshal(strings.ToLower(name))

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		if !ensurePineEditorOpen(ctx, c) {
			return fmt.Errorf("could not open Pine Editor")
		}
		expr := `(function() {
			var target = ` + string(escapedName) + `;
			return fetch('https://pine-facade.tradingview.com/pine-facade/list/?filter=saved', { credentials: 'include' })
				.then(function(r) { return r.json(); })
				.then(function(scripts) {
					if (!Array.isArray(scripts)) return {error: 'pine-facade returned unexpected data'};
					var match = null;
					for (var i = 0; i < scripts.length; i++) {
						var sn = (scripts[i].scriptName || '').toLowerCase();
						var st = (scripts[i].scriptTitle || '').toLowerCase();
						if (sn === target || st === target) { match = scripts[i]; break; }
					}
					if (!match) {
						for (var j = 0; j < scripts.length; j++) {
							var sn2 = (scripts[j].scriptName || '').toLowerCase();
							var st2 = (scripts[j].scriptTitle || '').toLowerCase();
							if (sn2.indexOf(target) !== -1 || st2.indexOf(target) !== -1) { match = scripts[j]; break; }
						}
					}
					if (!match) return {error: 'Script "' + target + '" not found. Use pine_list_scripts to see available scripts.'};
					var id = match.scriptIdPart;
					var ver = match.version || 1;
					return fetch('https://pine-facade.tradingview.com/pine-facade/get/' + id + '/' + ver, { credentials: 'include' })
						.then(function(r2) { return r2.json(); })
						.then(function(data) {
							var source = data.source || '';
							if (!source) return {error: 'Script source is empty', name: match.scriptName || match.scriptTitle};
							var m = ` + findMonaco + `;
							if (m) {
								m.editor.setValue(source);
								return {success: true, name: match.scriptName || match.scriptTitle, id: id, lines: source.split('\n').length};
							}
							return {error: 'Monaco editor not found to inject source', name: match.scriptName || match.scriptTitle};
						});
				})
				.catch(function(e) { return {error: e.message}; });
		})()`
		raw, err := c.Evaluate(ctx, expr, true)
		if err != nil {
			return err
		}
		var res map[string]interface{}
		if err := json.Unmarshal(raw, &res); err != nil {
			return fmt.Errorf("parse open result: %w", err)
		}
		if errMsg, ok := res["error"].(string); ok {
			return fmt.Errorf("%s", errMsg)
		}
		result = res
		result["source"] = "internal_api"
		result["opened"] = true
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// ListScripts returns all saved Pine Scripts from the pine-facade API.
func ListScripts() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	const expr = `fetch('https://pine-facade.tradingview.com/pine-facade/list/?filter=saved', { credentials: 'include' })
		.then(function(r) { return r.json(); })
		.then(function(data) {
			if (!Array.isArray(data)) return {scripts: [], error: 'Unexpected response from pine-facade'};
			return {
				scripts: data.map(function(s) {
					return {
						id: s.scriptIdPart || null,
						name: s.scriptName || s.scriptTitle || 'Untitled',
						title: s.scriptTitle || null,
						version: s.version || null,
						modified: s.modified || null,
					};
				})
			};
		})
		.catch(function(e) { return {scripts: [], error: e.message}; })`

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		raw, err := c.Evaluate(ctx, expr, true)
		if err != nil {
			return err
		}
		var res struct {
			Scripts []interface{} `json:"scripts"`
			Error   string        `json:"error"`
		}
		if err := json.Unmarshal(raw, &res); err != nil {
			return fmt.Errorf("parse list result: %w", err)
		}
		result = map[string]interface{}{
			"success": true,
			"scripts": res.Scripts,
			"count":   len(res.Scripts),
			"source":  "internal_api",
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

// Check sends source code to TradingView's server API for compilation validation.
// Does not require TradingView Desktop to be running.
func Check(source string) (map[string]interface{}, error) {
	const checkURL = "https://pine-facade.tradingview.com/pine-facade/translate_light?user_name=Guest&pine_id=00000000-0000-0000-0000-000000000000"

	body := url.Values{}
	body.Set("source", source)

	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodPost, checkURL,
		strings.NewReader(body.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "https://www.tradingview.com/")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("pine check request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TradingView API returned %d: %s", resp.StatusCode, resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var payload struct {
		Result *struct {
			Errors2   []struct {
				Start   *struct{ Line, Column int } `json:"start"`
				End     *struct{ Line, Column int } `json:"end"`
				Message string                      `json:"message"`
			} `json:"errors2"`
			Warnings2 []struct {
				Start   *struct{ Line, Column int } `json:"start"`
				Message string                      `json:"message"`
			} `json:"warnings2"`
		} `json:"result"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("parse check response: %w", err)
	}

	var errors []map[string]interface{}
	var warnings []map[string]interface{}

	if payload.Result != nil {
		for _, e := range payload.Result.Errors2 {
			m := map[string]interface{}{"message": e.Message}
			if e.Start != nil {
				m["line"] = e.Start.Line
				m["column"] = e.Start.Column
			}
			if e.End != nil {
				m["end_line"] = e.End.Line
				m["end_column"] = e.End.Column
			}
			errors = append(errors, m)
		}
		for _, w := range payload.Result.Warnings2 {
			m := map[string]interface{}{"message": w.Message}
			if w.Start != nil {
				m["line"] = w.Start.Line
				m["column"] = w.Start.Column
			}
			warnings = append(warnings, m)
		}
	}
	if payload.Error != "" {
		errors = append(errors, map[string]interface{}{"message": payload.Error})
	}

	compiled := len(errors) == 0
	result := map[string]interface{}{
		"success":       true,
		"compiled":      compiled,
		"error_count":   len(errors),
		"warning_count": len(warnings),
	}
	if len(errors) > 0 {
		result["errors"] = errors
	}
	if len(warnings) > 0 {
		result["warnings"] = warnings
	}
	if compiled {
		result["note"] = "Pine Script compiled successfully."
	}
	return result, nil
}

// RegisterTools registers all 12 Pine Script MCP tools.
func RegisterTools(reg *mcp.Registry) {
	reg.Register(mcp.ToolDef{
		Name:        "pine_get_source",
		Description: "Get current Pine Script source code from the editor. WARNING: can be very large for complex scripts.",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			return GetSource()
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "pine_set_source",
		Description: "Inject Pine Script source code into the editor",
		Schema: mcp.InputSchema{
			Type:       "object",
			Properties: map[string]mcp.PropertySchema{"source": {Type: "string", Description: "Pine Script source code to inject"}},
			Required:   []string{"source"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Source string `json:"source"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return SetSource(p.Source)
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "pine_compile",
		Description: "Compile / add the current Pine Script to the chart (clicks Add to chart button or Ctrl+Enter)",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			return Compile()
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "pine_smart_compile",
		Description: "Intelligent compile: detects button, compiles, checks errors, reports whether a new study was added",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			return SmartCompile()
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "pine_get_errors",
		Description: "Get Pine Script compilation errors from Monaco editor markers",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			return GetErrors()
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "pine_get_console",
		Description: "Read Pine Script console/log output (compile messages, log.info(), errors)",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			return GetConsole()
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "pine_save",
		Description: "Save the current Pine Script to TradingView cloud (Ctrl+S)",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			return Save()
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "pine_new",
		Description: "Create a new blank Pine Script of the specified type",
		Schema: mcp.InputSchema{
			Type:       "object",
			Properties: map[string]mcp.PropertySchema{"type": {Type: "string", Description: "indicator, strategy, or library"}},
			Required:   []string{"type"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Type string `json:"type"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return NewScript(p.Type)
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "pine_open",
		Description: "Open a saved Pine Script by name (case-insensitive, partial match supported)",
		Schema: mcp.InputSchema{
			Type:       "object",
			Properties: map[string]mcp.PropertySchema{"name": {Type: "string", Description: "Name of the saved script"}},
			Required:   []string{"name"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Name string `json:"name"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return OpenScript(p.Name)
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "pine_list_scripts",
		Description: "List all saved Pine Scripts from TradingView cloud",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			return ListScripts()
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "pine_analyze",
		Description: "Run static analysis on Pine Script code WITHOUT compiling — catches array out-of-bounds, unguarded array.first()/last(), and strategy misuse. Works offline, no TradingView connection needed.",
		Schema: mcp.InputSchema{
			Type:       "object",
			Properties: map[string]mcp.PropertySchema{"source": {Type: "string", Description: "Pine Script source code to analyze"}},
			Required:   []string{"source"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Source string `json:"source"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return Analyze(p.Source), nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "pine_check",
		Description: "Compile Pine Script via TradingView's server API without needing the chart open. Returns compile errors/warnings.",
		Schema: mcp.InputSchema{
			Type:       "object",
			Properties: map[string]mcp.PropertySchema{"source": {Type: "string", Description: "Pine Script source code to compile/validate"}},
			Required:   []string{"source"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Source string `json:"source"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			result, err := Check(p.Source)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})
}
