package cdp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// mockCDPServer starts a minimal CDP WebSocket server for testing.
// The handler receives each request Message and returns a response Message
// (ID is filled in automatically).
func mockCDPServer(t *testing.T, handler func(req Message) *Message) (*httptest.Server, *Target) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("upgrade: %v", err)
			return
		}
		defer conn.Close()
		var mu sync.Mutex
		for {
			var req Message
			if err := conn.ReadJSON(&req); err != nil {
				return
			}
			resp := handler(req)
			if resp != nil {
				resp.ID = req.ID
				mu.Lock()
				_ = conn.WriteJSON(resp)
				mu.Unlock()
			}
		}
	}))
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	target := &Target{
		ID:                   "mock-target",
		Type:                 "page",
		URL:                  "https://www.tradingview.com/chart/test/",
		WebSocketDebuggerURL: wsURL,
	}
	return srv, target
}

func TestClientEvaluateNumber(t *testing.T) {
	srv, target := mockCDPServer(t, func(req Message) *Message {
		return &Message{
			Result: json.RawMessage(`{"result":{"type":"number","value":42}}`),
		}
	})
	defer srv.Close()

	client, err := Connect(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	val, err := client.Evaluate(context.Background(), "42", false)
	if err != nil {
		t.Fatal(err)
	}
	var n float64
	if err := json.Unmarshal(val, &n); err != nil {
		t.Fatalf("unmarshal value: %v", err)
	}
	if n != 42 {
		t.Errorf("expected 42, got %v", n)
	}
}

func TestClientEnableDomains(t *testing.T) {
	enabled := map[string]bool{}
	var mu sync.Mutex
	srv, target := mockCDPServer(t, func(req Message) *Message {
		mu.Lock()
		enabled[req.Method] = true
		mu.Unlock()
		return &Message{Result: json.RawMessage(`{}`)}
	})
	defer srv.Close()

	client, err := Connect(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	if err := client.EnableDomains(context.Background()); err != nil {
		t.Fatal(err)
	}
	for _, m := range []string{"Runtime.enable", "Page.enable", "DOM.enable"} {
		mu.Lock()
		ok := enabled[m]
		mu.Unlock()
		if !ok {
			t.Errorf("domain %s was not enabled", m)
		}
	}
}

func TestClientLivenessCheck(t *testing.T) {
	srv, target := mockCDPServer(t, func(req Message) *Message {
		return &Message{Result: json.RawMessage(`{"result":{"type":"number","value":1}}`)}
	})
	defer srv.Close()

	client, err := Connect(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	if err := client.LivenessCheck(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestClientEvaluateJSError(t *testing.T) {
	srv, target := mockCDPServer(t, func(req Message) *Message {
		return &Message{
			Result: json.RawMessage(`{"result":{"type":"undefined"},"exceptionDetails":{"text":"ReferenceError","exception":{"description":"ReferenceError: x is not defined"}}}`),
		}
	})
	defer srv.Close()

	client, err := Connect(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	_, err = client.Evaluate(context.Background(), "x", false)
	if err == nil {
		t.Fatal("expected JS evaluation error, got nil")
	}
}

func TestClientCaptureScreenshot(t *testing.T) {
	srv, target := mockCDPServer(t, func(req Message) *Message {
		if req.Method == "Page.captureScreenshot" {
			return &Message{Result: json.RawMessage(`{"data":"aGVsbG8="}`)}
		}
		return &Message{Result: json.RawMessage(`{}`)}
	})
	defer srv.Close()

	client, err := Connect(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	data, err := client.CaptureScreenshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if data != "aGVsbG8=" {
		t.Errorf("expected base64 data, got %q", data)
	}
}
