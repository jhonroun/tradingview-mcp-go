package cdp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

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

func TestClientEvaluateWithOptionsAsyncValue(t *testing.T) {
	paramsCh := make(chan EvaluateParams, 1)
	srv, target := mockCDPServer(t, func(req Message) *Message {
		var params EvaluateParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			t.Errorf("unmarshal evaluate params: %v", err)
		}
		paramsCh <- params
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

	const expr = `(async () => 42)()`
	val, err := client.EvaluateWithOptions(context.Background(), expr, EvaluateOptions{
		AwaitPromise:  true,
		ReturnByValue: true,
		Timeout:       2500 * time.Millisecond,
	})
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

	select {
	case params := <-paramsCh:
		if params.Expression != expr {
			t.Errorf("expression = %q, want %q", params.Expression, expr)
		}
		if !params.AwaitPromise {
			t.Error("awaitPromise was not set")
		}
		if !params.ReturnByValue {
			t.Error("returnByValue was not set")
		}
		if params.Timeout != 2500 {
			t.Errorf("timeout = %d, want 2500", params.Timeout)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for evaluate params")
	}
}

func TestClientEvaluatePreservesMissingValue(t *testing.T) {
	srv, target := mockCDPServer(t, func(req Message) *Message {
		return &Message{
			Result: json.RawMessage(`{"result":{"type":"object","subtype":"promise","description":"Promise"}}`),
		}
	})
	defer srv.Close()

	client, err := Connect(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	val, err := client.Evaluate(context.Background(), "Promise.resolve(42)", false)
	if err != nil {
		t.Fatal(err)
	}
	if val != nil {
		t.Errorf("Evaluate returned %s, want nil missing value", string(val))
	}
}

func TestClientEvaluateWithOptionsCanReturnRemoteObject(t *testing.T) {
	srv, target := mockCDPServer(t, func(req Message) *Message {
		return &Message{
			Result: json.RawMessage(`{"result":{"type":"object","subtype":"promise","description":"Promise"}}`),
		}
	})
	defer srv.Close()

	client, err := Connect(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	val, err := client.EvaluateWithOptions(context.Background(), "Promise.resolve(42)", EvaluateOptions{
		ReturnByValue: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	var obj RemoteObject
	if err := json.Unmarshal(val, &obj); err != nil {
		t.Fatalf("unmarshal remote object: %v", err)
	}
	if obj.Type != "object" || obj.Subtype != "promise" {
		t.Errorf("remote object = %#v, want object/promise", obj)
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
