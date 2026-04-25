// Package replay implements replay_start, replay_step, replay_stop,
// replay_status, replay_autoplay, replay_trade.
package replay

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jhonroun/tradingview-mcp-go/internal/cdp"
	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
)

const replayAPI = `window.TradingViewApi._replayApi`

// validAutoplayDelays mirrors VALID_AUTOPLAY_DELAYS from core/replay.js.
var validAutoplayDelays = map[int]bool{
	100: true, 143: true, 200: true, 300: true,
	1000: true, 2000: true, 3000: true, 5000: true, 10000: true,
}

// wv unwraps a TradingView observable value in JS.
// mirrors the wv() helper in core/replay.js.
func wv(path string) string {
	return fmt.Sprintf(
		`(function(){ var v = %s; return (v && typeof v === 'object' && typeof v.value === 'function') ? v.value() : v; })()`,
		path,
	)
}

func evalBool(ctx context.Context, c *cdp.Client, expr string) (bool, error) {
	raw, err := c.Evaluate(ctx, expr, false)
	if err != nil {
		return false, err
	}
	var b bool
	if err := json.Unmarshal(raw, &b); err != nil {
		return false, nil
	}
	return b, nil
}

func evalAny(ctx context.Context, c *cdp.Client, expr string) (interface{}, error) {
	raw, err := c.Evaluate(ctx, expr, false)
	if err != nil {
		return nil, err
	}
	var v interface{}
	_ = json.Unmarshal(raw, &v)
	return v, nil
}

// Start enters replay mode, optionally at a specific date (YYYY-MM-DD).
func Start(date string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		rp := replayAPI

		// Check availability.
		avail, err := evalBool(ctx, c, wv(rp+".isReplayAvailable()"))
		if err != nil {
			return err
		}
		if !avail {
			return fmt.Errorf("replay is not available for the current symbol/timeframe")
		}

		// Show toolbar.
		if _, err := c.Evaluate(ctx, rp+".showReplayToolbar()", false); err != nil {
			return err
		}

		// Select date or first available.
		if date != "" {
			// Parse to timestamp (ms).
			t, err := time.Parse("2006-01-02", date)
			if err != nil {
				return fmt.Errorf("invalid date %q; use YYYY-MM-DD format", date)
			}
			tsMs := t.UnixMilli()
			expr := fmt.Sprintf(`%s.selectDate(%d).then(function() { return 'ok'; })`, rp, tsMs)
			if _, err := c.Evaluate(ctx, expr, true); err != nil {
				return err
			}
		} else {
			if _, err := c.Evaluate(ctx, rp+".selectFirstAvailableDate()", false); err != nil {
				return err
			}
		}

		// Poll until isReplayStarted AND currentDate is set (up to 30×250 ms = 7.5 s).
		var started bool
		var currentDate interface{}
		for i := 0; i < 30; i++ {
			started, _ = evalBool(ctx, c, wv(rp+".isReplayStarted()"))
			currentDate, _ = evalAny(ctx, c, wv(rp+".currentDate()"))
			if started && currentDate != nil {
				break
			}
			time.Sleep(250 * time.Millisecond)
		}

		if !started {
			_, _ = c.Evaluate(ctx, rp+".stopReplay()", false)
			return fmt.Errorf("replay failed to start. The selected date may not have data for this timeframe. Try a more recent date or a higher timeframe (e.g., Daily)")
		}

		startedFrom := date
		if startedFrom == "" {
			startedFrom = "(first available)"
		}
		result = map[string]interface{}{
			"success":       true,
			"replay_started": true,
			"date":          startedFrom,
			"current_date":  currentDate,
		}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// Step advances one bar in replay mode.
func Step() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		rp := replayAPI

		started, err := evalBool(ctx, c, wv(rp+".isReplayStarted()"))
		if err != nil {
			return err
		}
		if !started {
			return fmt.Errorf("replay is not started. Use replay_start first")
		}

		before, _ := evalAny(ctx, c, wv(rp+".currentDate()"))
		if _, err := c.Evaluate(ctx, rp+".doStep()", false); err != nil {
			return err
		}

		// Poll up to 12×250 ms = 3 s for currentDate to change.
		currentDate := before
		for i := 0; i < 12; i++ {
			time.Sleep(250 * time.Millisecond)
			currentDate, _ = evalAny(ctx, c, wv(rp+".currentDate()"))
			if currentDate != before {
				break
			}
		}

		result = map[string]interface{}{
			"success":      true,
			"action":       "step",
			"current_date": currentDate,
		}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// Stop exits replay mode and returns to realtime.
func Stop() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		rp := replayAPI

		started, err := evalBool(ctx, c, wv(rp+".isReplayStarted()"))
		if err != nil {
			return err
		}
		if !started {
			result = map[string]interface{}{"success": true, "action": "already_stopped"}
			return nil
		}

		if _, err := c.Evaluate(ctx, rp+".stopReplay()", false); err != nil {
			return err
		}
		result = map[string]interface{}{"success": true, "action": "replay_stopped"}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// Status returns the current replay mode state.
func Status() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		rp := replayAPI

		const stExpr = `(function() {
			var r = ` + replayAPI + `;
			function unwrap(v) { return (v && typeof v === 'object' && typeof v.value === 'function') ? v.value() : v; }
			return {
				is_replay_available:  unwrap(r.isReplayAvailable()),
				is_replay_started:    unwrap(r.isReplayStarted()),
				is_autoplay_started:  unwrap(r.isAutoplayStarted()),
				replay_mode:          unwrap(r.replayMode()),
				current_date:         unwrap(r.currentDate()),
				autoplay_delay:       unwrap(r.autoplayDelay()),
			};
		})()`

		raw, err := c.Evaluate(ctx, stExpr, false)
		if err != nil {
			return err
		}
		var st map[string]interface{}
		if err := json.Unmarshal(raw, &st); err != nil {
			return fmt.Errorf("parse status: %w", err)
		}

		pos, _ := evalAny(ctx, c, wv(rp+".position()"))
		pnl, _ := evalAny(ctx, c, wv(rp+".realizedPL()"))

		result = map[string]interface{}{"success": true}
		for k, v := range st {
			result[k] = v
		}
		result["position"] = pos
		result["realized_pnl"] = pnl
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// Autoplay toggles autoplay in replay mode, optionally changing speed (0 = just toggle).
func Autoplay(speedMs int) (map[string]interface{}, error) {
	// Validate before any CDP calls — mirrors Node.js validation order.
	if speedMs > 0 && !validAutoplayDelays[speedMs] {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("invalid autoplay delay %dms. Valid values: 100, 143, 200, 300, 1000, 2000, 3000, 5000, 10000", speedMs),
		}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		rp := replayAPI

		started, err := evalBool(ctx, c, wv(rp+".isReplayStarted()"))
		if err != nil {
			return err
		}
		if !started {
			return fmt.Errorf("replay is not started. Use replay_start first")
		}

		if speedMs > 0 {
			if _, err := c.Evaluate(ctx, fmt.Sprintf(`%s.changeAutoplayDelay(%d)`, rp, speedMs), false); err != nil {
				return err
			}
		}
		if _, err := c.Evaluate(ctx, rp+".toggleAutoplay()", false); err != nil {
			return err
		}

		isAutoplay, _ := evalAny(ctx, c, wv(rp+".isAutoplayStarted()"))
		delay, _ := evalAny(ctx, c, wv(rp+".autoplayDelay()"))

		autoplayActive := false
		if b, ok := isAutoplay.(bool); ok {
			autoplayActive = b
		}
		result = map[string]interface{}{
			"success":        true,
			"autoplay_active": autoplayActive,
			"delay_ms":       delay,
		}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// Trade executes buy, sell, or close in replay mode.
func Trade(action string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var result map[string]interface{}
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		rp := replayAPI

		started, err := evalBool(ctx, c, wv(rp+".isReplayStarted()"))
		if err != nil {
			return err
		}
		if !started {
			return fmt.Errorf("replay is not started. Use replay_start first")
		}

		var tradeExpr string
		switch action {
		case "buy":
			tradeExpr = rp + ".buy()"
		case "sell":
			tradeExpr = rp + ".sell()"
		case "close":
			tradeExpr = rp + ".closePosition()"
		default:
			return fmt.Errorf("invalid action. Use: buy, sell, or close")
		}

		if _, err := c.Evaluate(ctx, tradeExpr, false); err != nil {
			return err
		}

		pos, _ := evalAny(ctx, c, wv(rp+".position()"))
		pnl, _ := evalAny(ctx, c, wv(rp+".realizedPL()"))

		result = map[string]interface{}{
			"success":      true,
			"action":       action,
			"position":     pos,
			"realized_pnl": pnl,
		}
		return nil
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return result, nil
}

// RegisterTools registers replay_start, replay_step, replay_stop,
// replay_status, replay_autoplay, replay_trade.
func RegisterTools(reg *mcp.Registry) {
	reg.Register(mcp.ToolDef{
		Name:        "replay_start",
		Description: "Start bar replay mode, optionally at a specific date",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"date": {Type: "string", Description: "Date to start replay from (YYYY-MM-DD format). If omitted, selects first available date."},
			},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Date string `json:"date"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return Start(p.Date)
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "replay_step",
		Description: "Advance one bar in replay mode",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			return Step()
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "replay_stop",
		Description: "Stop replay and return to realtime",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			return Stop()
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "replay_status",
		Description: "Get current replay mode status including position and P&L",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			return Status()
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "replay_autoplay",
		Description: "Toggle autoplay in replay mode, optionally set speed",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"speed": {Type: "number", Description: "Autoplay delay in ms (lower = faster). Valid values: 100, 143, 200, 300, 1000, 2000, 3000, 5000, 10000. Leave empty to just toggle."},
			},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Speed int `json:"speed"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return Autoplay(p.Speed)
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "replay_trade",
		Description: "Execute a trade action in replay mode (buy, sell, or close position)",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"action": {Type: "string", Description: "Trade action: buy, sell, or close"},
			},
			Required: []string{"action"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Action string `json:"action"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return Trade(p.Action)
		},
	})
}
