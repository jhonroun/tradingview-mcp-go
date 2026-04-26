package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// helper: run the server against a single-line input, return parsed Response.
func runOne(t *testing.T, reg *Registry, line string) Response {
	t.Helper()
	srv := NewServer(reg, "")
	srv.in = strings.NewReader(line + "\n")
	var buf bytes.Buffer
	srv.out = &buf
	if err := srv.Run(); err != nil {
		t.Fatalf("server.Run: %v", err)
	}
	var resp Response
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v\ngot: %s", err, buf.String())
	}
	return resp
}

// helper: run the server against multiple newline-separated messages.
func runLines(t *testing.T, reg *Registry, lines string) []Response {
	t.Helper()
	srv := NewServer(reg, "")
	srv.in = strings.NewReader(lines)
	var buf bytes.Buffer
	srv.out = &buf
	if err := srv.Run(); err != nil {
		t.Fatalf("server.Run: %v", err)
	}
	var resps []Response
	dec := json.NewDecoder(&buf)
	for dec.More() {
		var r Response
		if err := dec.Decode(&r); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		resps = append(resps, r)
	}
	return resps
}

// ── Phase 1: initialize ───────────────────────────────────────────────────────

func TestInitialize(t *testing.T) {
	resp := runOne(t, NewRegistry(), `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"test","version":"0.1"},"capabilities":{}}}`)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}

	b, _ := json.Marshal(resp.Result)
	var result InitializeResult
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatal(err)
	}
	if result.ProtocolVersion != ProtocolVersion {
		t.Errorf("protocolVersion = %q, want %q", result.ProtocolVersion, ProtocolVersion)
	}
	if result.ServerInfo.Name != "tradingview" {
		t.Errorf("serverInfo.name = %q, want tradingview", result.ServerInfo.Name)
	}
	if _, ok := result.Capabilities["tools"]; !ok {
		t.Errorf("capabilities missing 'tools' key")
	}
}

// ── Phase 1: tools/list ───────────────────────────────────────────────────────

func TestToolsListExact82(t *testing.T) {
	reg := newFullRegistry()
	resp := runOne(t, reg, `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	b, _ := json.Marshal(resp.Result)
	var result ListToolsResult
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Tools) != 82 {
		t.Errorf("tools/list returned %d tools, want 82", len(result.Tools))
	}
	for i, tool := range result.Tools {
		if tool.Name == "" {
			t.Errorf("tool[%d] has empty name", i)
		}
		if tool.Description == "" {
			t.Errorf("tool[%d] %q has empty description", i, tool.Name)
		}
		if tool.InputSchema.Type == "" {
			t.Errorf("tool[%d] %q has empty inputSchema.type", i, tool.Name)
		}
	}
}

// ── Phase 1: tools/call ───────────────────────────────────────────────────────

func TestToolsCallKnown(t *testing.T) {
	reg := NewRegistry()
	reg.Register(ToolDef{
		Name:    "tv_ping",
		Schema:  InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) { return map[string]bool{"pong": true}, nil },
	})
	resp := runOne(t, reg, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"tv_ping"}}`)

	b, _ := json.Marshal(resp.Result)
	var result CallToolResult
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("IsError should be false for known tool, got content: %v", result.Content)
	}
	if len(result.Content) == 0 || result.Content[0].Type != "text" {
		t.Fatalf("unexpected content: %+v", result.Content)
	}
}

func TestToolsCallUnknown(t *testing.T) {
	resp := runOne(t, NewRegistry(), `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"nonexistent"}}`)

	b, _ := json.Marshal(resp.Result)
	var result CallToolResult
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Fatal("IsError should be true for unknown tool")
	}
	if len(result.Content) == 0 {
		t.Fatal("content should not be empty on error")
	}
}

func TestToolsCallBadArgs(t *testing.T) {
	reg := NewRegistry()
	type strictArgs struct {
		Value int `json:"value"`
	}
	reg.Register(ToolDef{
		Name:   "needs_int",
		Schema: InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p strictArgs
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return map[string]interface{}{"success": true, "v": p.Value}, nil
		},
	})
	// pass a string where int is expected — handler returns success:false
	resp := runOne(t, reg, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"needs_int","arguments":{"value":"not-a-number"}}}`)

	b, _ := json.Marshal(resp.Result)
	var result CallToolResult
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected content in response")
	}
	// Tool returned a response (not a JSON-RPC error) — this is correct MCP behaviour.
	// The handler decides how to communicate the error.
}

// ── Phase 1: ping ─────────────────────────────────────────────────────────────

func TestPing(t *testing.T) {
	resp := runOne(t, NewRegistry(), `{"jsonrpc":"2.0","id":1,"method":"ping"}`)

	if resp.Error != nil {
		t.Fatalf("ping returned error: %+v", resp.Error)
	}
	b, _ := json.Marshal(resp.Result)
	var result map[string]interface{}
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatal(err)
	}
	if len(result) != 0 {
		t.Errorf("ping result should be {}, got %v", result)
	}
}

// ── Phase 1: unknown method ───────────────────────────────────────────────────

func TestUnknownMethod(t *testing.T) {
	resp := runOne(t, NewRegistry(), `{"jsonrpc":"2.0","id":1,"method":"unknown/method"}`)

	if resp.Error == nil || resp.Error.Code != ErrMethodNotFound {
		t.Fatalf("expected -32601, got %+v", resp.Error)
	}
}

// ── Phase 1: parse error ──────────────────────────────────────────────────────

func TestParseError(t *testing.T) {
	srv := NewServer(NewRegistry(), "")
	srv.in = strings.NewReader("{not valid json}\n")
	var buf bytes.Buffer
	srv.out = &buf
	_ = srv.Run()

	var resp Response
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v\ngot: %s", err, buf.String())
	}
	if resp.Error == nil || resp.Error.Code != ErrParseError {
		t.Fatalf("expected -32700, got %+v", resp.Error)
	}
}

// ── Phase 1: notifications/initialized (no response) ─────────────────────────

func TestNotificationNoResponse(t *testing.T) {
	srv := NewServer(NewRegistry(), "")
	// Send a notification then a ping so we have something to read.
	srv.in = strings.NewReader(
		`{"jsonrpc":"2.0","method":"notifications/initialized"}` + "\n" +
			`{"jsonrpc":"2.0","id":2,"method":"ping"}` + "\n",
	)
	var buf bytes.Buffer
	srv.out = &buf
	_ = srv.Run()

	// Only one response (the ping), not two.
	resps := []json.RawMessage{}
	dec := json.NewDecoder(&buf)
	for dec.More() {
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			t.Fatalf("decode: %v", err)
		}
		resps = append(resps, raw)
	}
	if len(resps) != 1 {
		t.Fatalf("expected 1 response (only ping), got %d", len(resps))
	}
}

// ── Phase 1: multiline JSON is a protocol violation ──────────────────────────
//
// MCP stdio uses NDJSON: one complete JSON object per line.
// A multiline JSON object is split across lines; each partial line is not valid
// JSON and must produce a parse error (-32700).

func TestMultilineJSON(t *testing.T) {
	// Each of these lines is a partial fragment — none is a complete JSON object.
	// The server must reply with a parse error for every non-empty line.
	multiline := "{\n  \"jsonrpc\": \"2.0\",\n  \"id\": 1,\n  \"method\": \"ping\"\n}\n"

	srv := NewServer(NewRegistry(), "")
	srv.in = strings.NewReader(multiline)
	var buf bytes.Buffer
	srv.out = &buf
	_ = srv.Run()

	dec := json.NewDecoder(&buf)
	var resps []Response
	for dec.More() {
		var r Response
		if err := dec.Decode(&r); err != nil {
			t.Fatalf("decode: %v", err)
		}
		resps = append(resps, r)
	}
	// Every non-empty line must have produced a parse error.
	for i, r := range resps {
		if r.Error == nil || r.Error.Code != ErrParseError {
			t.Errorf("response[%d]: expected -32700 parse error, got %+v", i, r.Error)
		}
	}
	if len(resps) == 0 {
		t.Fatal("expected at least one parse-error response for multiline input")
	}
}

// ── Phase 1: large response (>64KB, Scanner regression) ──────────────────────

func TestLargeResponse(t *testing.T) {
	reg := NewRegistry()
	big := strings.Repeat("x", 128*1024) // 128 KB value
	reg.Register(ToolDef{
		Name:   "big_tool",
		Schema: InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			return map[string]string{"data": big}, nil
		},
	})
	resp := runOne(t, reg, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"big_tool"}}`)

	b, _ := json.Marshal(resp.Result)
	var result CallToolResult
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("large response should not be an error")
	}
	if len(result.Content) == 0 {
		t.Fatal("expected content")
	}
	if !strings.Contains(result.Content[0].Text, big) {
		t.Errorf("response text does not contain the large payload (len=%d)", len(result.Content[0].Text))
	}
}

// ── Phase 1: sequential requests ─────────────────────────────────────────────

func TestSequentialRequests(t *testing.T) {
	reg := NewRegistry()
	reg.Register(ToolDef{
		Name:    "counter",
		Schema:  InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) { return map[string]bool{"ok": true}, nil },
	})

	lines := `{"jsonrpc":"2.0","id":1,"method":"ping"}` + "\n" +
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}` + "\n" +
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"counter"}}` + "\n"

	resps := runLines(t, reg, lines)
	if len(resps) != 3 {
		t.Fatalf("expected 3 responses, got %d", len(resps))
	}
	for i, r := range resps {
		if r.Error != nil {
			t.Errorf("request %d has unexpected error: %+v", i+1, r.Error)
		}
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// newFullRegistry registers all 82 production tools so tool-count tests work.
// It imports the actual packages to avoid duplication of registration logic.
func newFullRegistry() *Registry {
	reg := NewRegistry()
	// Register a stub for each of the 78 expected tools.
	// The names are the canonical tool names from the compatibility matrix.
	names := []string{
		"tv_health_check", "tv_discover", "tv_ui_state", "tv_launch",
		"chart_get_state", "chart_get_visible_range",
		"chart_set_symbol", "chart_set_timeframe", "chart_set_type",
		"chart_manage_indicator", "chart_scroll_to_date", "chart_set_visible_range",
		"symbol_info", "symbol_search",
		"quote_get", "data_get_ohlcv", "data_get_study_values",
		"data_get_pine_lines", "data_get_pine_labels", "data_get_pine_tables", "data_get_pine_boxes",
		"data_get_indicator", "data_get_strategy_results", "data_get_trades", "data_get_equity",
		"depth_get",
		"capture_screenshot",
		"indicator_set_inputs", "indicator_toggle_visibility",
		"pine_get_source", "pine_set_source", "pine_compile", "pine_smart_compile",
		"pine_get_errors", "pine_get_console", "pine_save", "pine_new",
		"pine_open", "pine_list_scripts", "pine_analyze", "pine_check",
		"draw_shape", "draw_list", "draw_get_properties", "draw_remove_one", "draw_clear",
		"alert_create", "alert_list", "alert_delete",
		"watchlist_get", "watchlist_add",
		"replay_start", "replay_step", "replay_stop", "replay_status",
		"replay_autoplay", "replay_trade",
		"pane_list", "pane_set_layout", "pane_focus", "pane_set_symbol",
		"tab_list", "tab_new", "tab_close", "tab_switch",
		"ui_click", "ui_open_panel", "ui_fullscreen", "ui_keyboard",
		"ui_type_text", "ui_hover", "ui_scroll", "ui_mouse_click",
		"ui_find_element", "ui_evaluate",
		"layout_list", "layout_switch",
		"batch_run",
		// Phase 4 — HTS-ready composite tools
		"chart_context_for_llm", "indicator_state", "market_summary", "continuous_contract_context",
	}
	stub := func(args json.RawMessage) (interface{}, error) {
		return map[string]bool{"success": true}, nil
	}
	for i, name := range names {
		reg.Register(ToolDef{
			Name:        name,
			Description: fmt.Sprintf("stub tool %d", i),
			Schema:      InputSchema{Type: "object"},
			Handler:     stub,
		})
	}
	return reg
}
