// Package tab implements tab_list, tab_new, tab_close, tab_switch.
package tab

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jhonroun/tradingview-mcp-go/internal/cdp"
	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
)

const (
	cdpHost = "localhost"
	cdpPort = 9222
)

// ListTabs returns all TradingView chart tabs (pages) visible to the debug port.
func ListTabs() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	targets, err := cdp.ListTargets(ctx, cdpHost, cdpPort)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}

	var tabs []map[string]interface{}
	for _, t := range targets {
		if t.Type == "page" && strings.Contains(strings.ToLower(t.URL), "tradingview") {
			tabs = append(tabs, map[string]interface{}{
				"id":    t.ID,
				"title": t.Title,
				"url":   t.URL,
			})
		}
	}
	return map[string]interface{}{
		"success":   true,
		"tab_count": len(tabs),
		"tabs":      tabs,
	}, nil
}

// NewTab opens a new TradingView tab using Ctrl+T in the active chart window.
func NewTab() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Ctrl+T: modifiers=2 (Ctrl on Windows/Linux), keyCode=84 ('T')
	keyParams := cdp.KeyEventParams{
		Type:                  "keyDown",
		Key:                   "t",
		Code:                  "KeyT",
		Modifiers:             2,
		WindowsVirtualKeyCode: 84,
	}

	var beforeIDs []string
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		targets, err := cdp.ListTargets(ctx, cdpHost, cdpPort)
		if err != nil {
			return err
		}
		for _, t := range targets {
			if t.Type == "page" && strings.Contains(strings.ToLower(t.URL), "tradingview") {
				beforeIDs = append(beforeIDs, t.ID)
			}
		}
		if err := c.DispatchKeyEvent(ctx, keyParams); err != nil {
			return err
		}
		keyParams.Type = "keyUp"
		return c.DispatchKeyEvent(ctx, keyParams)
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}

	time.Sleep(1 * time.Second)
	result, _ := ListTabs()
	if result == nil {
		result = map[string]interface{}{}
	}
	result["success"] = true
	return result, nil
}

// CloseTab closes the currently active chart tab using Ctrl+W.
// Refuses to close if only one chart tab remains.
func CloseTab() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check we have ≥2 tabs before closing.
	targets, err := cdp.ListTargets(ctx, cdpHost, cdpPort)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	var tvTabs []cdp.Target
	for _, t := range targets {
		if t.Type == "page" && strings.Contains(strings.ToLower(t.URL), "tradingview") {
			tvTabs = append(tvTabs, t)
		}
	}
	if len(tvTabs) < 2 {
		return map[string]interface{}{
			"success": false,
			"error":   "cannot close the last chart tab",
		}, nil
	}

	keyParams := cdp.KeyEventParams{
		Type:                  "keyDown",
		Key:                   "w",
		Code:                  "KeyW",
		Modifiers:             2,
		WindowsVirtualKeyCode: 87,
	}

	err = cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		if err := c.DispatchKeyEvent(ctx, keyParams); err != nil {
			return err
		}
		keyParams.Type = "keyUp"
		return c.DispatchKeyEvent(ctx, keyParams)
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}

	time.Sleep(500 * time.Millisecond)
	result, _ := ListTabs()
	if result == nil {
		result = map[string]interface{}{}
	}
	result["success"] = true
	return result, nil
}

// SwitchTab activates a specific tab by its CDP target ID using /json/activate.
func SwitchTab(tabID string) (map[string]interface{}, error) {
	if tabID == "" {
		return map[string]interface{}{"success": false, "error": "tab_id is required"}, nil
	}

	url := fmt.Sprintf("http://%s:%d/json/activate/%s", cdpHost, cdpPort, tabID)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("activate failed (status %d): %s", resp.StatusCode, strings.TrimSpace(string(body))),
		}, nil
	}
	return map[string]interface{}{"success": true, "activated_tab_id": tabID}, nil
}

// RegisterTools registers tab_list, tab_new, tab_close, tab_switch.
func RegisterTools(reg *mcp.Registry) {
	reg.Register(mcp.ToolDef{
		Name:        "tab_list",
		Description: "List all open TradingView chart tabs with their IDs and titles",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			return ListTabs()
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "tab_new",
		Description: "Open a new TradingView chart tab (sends Ctrl+T to the active window)",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			return NewTab()
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "tab_close",
		Description: "Close the currently active chart tab. Refuses if only one tab remains.",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			return CloseTab()
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "tab_switch",
		Description: "Switch to a specific chart tab by its CDP target ID (from tab_list)",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"tab_id": {Type: "string", Description: "CDP target ID of the tab to activate"},
			},
			Required: []string{"tab_id"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				TabID string `json:"tab_id"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return SwitchTab(p.TabID)
		},
	})
}
