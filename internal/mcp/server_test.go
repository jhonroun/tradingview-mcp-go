package mcp

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestServerInitialize(t *testing.T) {
	reg := NewRegistry()
	srv := NewServer(reg, "test instructions")
	srv.in = strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"test","version":"0.1"},"capabilities":{}}}` + "\n")
	var buf bytes.Buffer
	srv.out = &buf

	if err := srv.Run(); err != nil {
		t.Fatal(err)
	}

	var resp Response
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v\ngot: %s", err, buf.String())
	}
	if resp.Error != nil {
		t.Fatalf("expected no error, got %+v", resp.Error)
	}

	var result InitializeResult
	b, _ := json.Marshal(resp.Result)
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatal(err)
	}
	if result.ProtocolVersion != ProtocolVersion {
		t.Errorf("protocolVersion = %q, want %q", result.ProtocolVersion, ProtocolVersion)
	}
	if result.ServerInfo.Name != "tradingview" {
		t.Errorf("serverInfo.name = %q, want tradingview", result.ServerInfo.Name)
	}
}

func TestServerListTools(t *testing.T) {
	reg := NewRegistry()
	reg.Register(ToolDef{
		Name:        "tv_test",
		Description: "test tool",
		Schema:      InputSchema{Type: "object"},
		Handler:     func(args json.RawMessage) (interface{}, error) { return map[string]bool{"ok": true}, nil },
	})
	srv := NewServer(reg, "")
	srv.in = strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}` + "\n")
	var buf bytes.Buffer
	srv.out = &buf

	if err := srv.Run(); err != nil {
		t.Fatal(err)
	}

	var resp struct {
		Result ListToolsResult `json:"result"`
	}
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v\ngot: %s", err, buf.String())
	}
	if len(resp.Result.Tools) != 1 || resp.Result.Tools[0].Name != "tv_test" {
		t.Fatalf("unexpected tools: %+v", resp.Result.Tools)
	}
}

func TestServerCallTool(t *testing.T) {
	reg := NewRegistry()
	reg.Register(ToolDef{
		Name:        "tv_ping",
		Description: "ping",
		Schema:      InputSchema{Type: "object"},
		Handler:     func(args json.RawMessage) (interface{}, error) { return map[string]bool{"pong": true}, nil },
	})
	srv := NewServer(reg, "")
	srv.in = strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"tv_ping"}}` + "\n")
	var buf bytes.Buffer
	srv.out = &buf

	if err := srv.Run(); err != nil {
		t.Fatal(err)
	}

	var resp struct {
		Result CallToolResult `json:"result"`
	}
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v\ngot: %s", err, buf.String())
	}
	if resp.Result.IsError {
		t.Fatalf("unexpected tool error: %v", resp.Result.Content)
	}
	if len(resp.Result.Content) == 0 || resp.Result.Content[0].Type != "text" {
		t.Fatalf("unexpected content: %+v", resp.Result.Content)
	}
}

func TestServerUnknownMethod(t *testing.T) {
	srv := NewServer(NewRegistry(), "")
	srv.in = strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"unknown/method"}` + "\n")
	var buf bytes.Buffer
	srv.out = &buf

	_ = srv.Run()

	var resp Response
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Error == nil || resp.Error.Code != ErrMethodNotFound {
		t.Fatalf("expected method-not-found error, got %+v", resp.Error)
	}
}

func TestRegistryCallUnknown(t *testing.T) {
	reg := NewRegistry()
	_, err := reg.Call("nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
}
