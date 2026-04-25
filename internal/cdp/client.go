package cdp

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	defaultHost = "localhost"
	defaultPort = 9222
	maxRetries  = 5
	baseDelay   = 500 * time.Millisecond
)

// Client is a CDP WebSocket client connected to one target.
type Client struct {
	conn    *websocket.Conn
	mu      sync.Mutex // serialises writes and pending map access
	pending map[int]chan Message
	nextID  atomic.Int32
}

// Connect establishes a CDP WebSocket connection to the given target.
func Connect(ctx context.Context, target *Target) (*Client, error) {
	dialer := websocket.Dialer{HandshakeTimeout: 10 * time.Second}
	conn, _, err := dialer.DialContext(ctx, target.WebSocketDebuggerURL, http.Header{})
	if err != nil {
		return nil, fmt.Errorf("websocket dial %s: %w", target.WebSocketDebuggerURL, err)
	}
	c := &Client{conn: conn, pending: make(map[int]chan Message)}
	go c.readLoop()
	return c, nil
}

// ConnectWithRetry discovers a TradingView target and connects, retrying with backoff.
func ConnectWithRetry(ctx context.Context) (*Client, *Target, error) {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		targets, err := ListTargets(ctx, defaultHost, defaultPort)
		if err == nil {
			if target, err := FindChartTarget(targets); err == nil {
				if client, err := Connect(ctx, target); err == nil {
					return client, target, nil
				} else {
					lastErr = err
				}
			} else {
				lastErr = err
			}
		} else {
			lastErr = err
		}
		delay := time.Duration(math.Min(
			float64(baseDelay)*math.Pow(2, float64(attempt)),
			float64(30*time.Second),
		))
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-time.After(delay):
		}
	}
	return nil, nil, fmt.Errorf("CDP connection failed after %d attempts: %w", maxRetries, lastErr)
}

// EnableDomains enables Runtime, Page, and DOM CDP domains.
func (c *Client) EnableDomains(ctx context.Context) error {
	for _, method := range []string{"Runtime.enable", "Page.enable", "DOM.enable"} {
		if _, err := c.call(ctx, method, nil); err != nil {
			return fmt.Errorf("%s: %w", method, err)
		}
	}
	return nil
}

// Evaluate executes a JavaScript expression and returns the raw JSON value.
func (c *Client) Evaluate(ctx context.Context, expression string, awaitPromise bool) (json.RawMessage, error) {
	raw, err := c.call(ctx, "Runtime.evaluate", EvaluateParams{
		Expression:    expression,
		ReturnByValue: true,
		AwaitPromise:  awaitPromise,
	})
	if err != nil {
		return nil, err
	}
	var result EvaluateResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse evaluate result: %w", err)
	}
	if result.ExceptionDetails != nil {
		msg := result.ExceptionDetails.Text
		if result.ExceptionDetails.Exception != nil {
			msg = result.ExceptionDetails.Exception.Description
		}
		return nil, fmt.Errorf("JS evaluation error: %s", msg)
	}
	return result.Result.Value, nil
}

// ScreenshotClip defines a viewport region for a clipped screenshot.
type ScreenshotClip struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Scale  float64 `json:"scale"`
}

// CaptureScreenshot captures a full-page PNG screenshot as base64-encoded data.
func (c *Client) CaptureScreenshot(ctx context.Context) (string, error) {
	return c.screenshot(ctx, map[string]interface{}{"format": "png"})
}

// CaptureScreenshotClip captures a clipped PNG screenshot as base64-encoded data.
func (c *Client) CaptureScreenshotClip(ctx context.Context, clip ScreenshotClip) (string, error) {
	return c.screenshot(ctx, map[string]interface{}{"format": "png", "clip": clip})
}

func (c *Client) screenshot(ctx context.Context, params map[string]interface{}) (string, error) {
	raw, err := c.call(ctx, "Page.captureScreenshot", params)
	if err != nil {
		return "", err
	}
	var result struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("parse screenshot result: %w", err)
	}
	return result.Data, nil
}

// KeyEventParams holds parameters for Input.dispatchKeyEvent.
type KeyEventParams struct {
	Type                  string `json:"type"`
	Key                   string `json:"key"`
	Code                  string `json:"code"`
	Modifiers             int    `json:"modifiers,omitempty"`
	WindowsVirtualKeyCode int    `json:"windowsVirtualKeyCode,omitempty"`
}

// DispatchKeyEvent sends a synthetic key event to the focused element.
func (c *Client) DispatchKeyEvent(ctx context.Context, p KeyEventParams) error {
	_, err := c.call(ctx, "Input.dispatchKeyEvent", p)
	return err
}

// InsertText inserts text at the current focus point (Input.insertText).
func (c *Client) InsertText(ctx context.Context, text string) error {
	_, err := c.call(ctx, "Input.insertText", map[string]interface{}{"text": text})
	return err
}

// MouseEventParams holds parameters for Input.dispatchMouseEvent.
type MouseEventParams struct {
	Type       string  `json:"type"`
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Button     string  `json:"button,omitempty"`
	Buttons    int     `json:"buttons,omitempty"`
	ClickCount int     `json:"clickCount,omitempty"`
	DeltaX     float64 `json:"deltaX,omitempty"`
	DeltaY     float64 `json:"deltaY,omitempty"`
}

// DispatchMouseEvent sends a synthetic mouse event (Input.dispatchMouseEvent).
func (c *Client) DispatchMouseEvent(ctx context.Context, p MouseEventParams) error {
	_, err := c.call(ctx, "Input.dispatchMouseEvent", p)
	return err
}

// LivenessCheck verifies the CDP connection is still alive.
func (c *Client) LivenessCheck(ctx context.Context) error {
	_, err := c.Evaluate(ctx, "1", false)
	return err
}

// Close closes the underlying WebSocket connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) call(ctx context.Context, method string, params interface{}) (json.RawMessage, error) {
	id := int(c.nextID.Add(1))
	ch := make(chan Message, 1)

	msg := map[string]interface{}{"id": id, "method": method}
	if params != nil {
		msg["params"] = params
	}

	c.mu.Lock()
	c.pending[id] = ch
	err := c.conn.WriteJSON(msg)
	if err != nil {
		delete(c.pending, id)
	}
	c.mu.Unlock()

	if err != nil {
		return nil, fmt.Errorf("write CDP message: %w", err)
	}

	select {
	case resp := <-ch:
		if resp.Error != nil {
			return nil, fmt.Errorf("CDP error %d: %s", resp.Error.Code, resp.Error.Message)
		}
		return resp.Result, nil
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, ctx.Err()
	}
}

func (c *Client) readLoop() {
	for {
		var msg Message
		if err := c.conn.ReadJSON(&msg); err != nil {
			return // connection closed
		}
		if msg.ID == 0 {
			continue // CDP event (no ID)
		}
		c.mu.Lock()
		ch, ok := c.pending[msg.ID]
		if ok {
			delete(c.pending, msg.ID)
		}
		c.mu.Unlock()
		if ok {
			ch <- msg
		}
	}
}
