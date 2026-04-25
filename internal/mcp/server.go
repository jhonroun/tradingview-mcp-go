package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const serverVersion = "2.0.0"

// Server is the MCP stdio server.
type Server struct {
	registry     *Registry
	instructions string
	in           io.Reader
	out          io.Writer
}

// NewServer creates a new MCP server backed by the given registry.
func NewServer(registry *Registry, instructions string) *Server {
	return &Server{
		registry:     registry,
		instructions: instructions,
		in:           os.Stdin,
		out:          os.Stdout,
	}
}

// Run starts the JSON-RPC read loop over stdin/stdout.
func (s *Server) Run() error {
	dec := json.NewDecoder(s.in)
	enc := json.NewEncoder(s.out)

	for {
		var req Request
		if err := dec.Decode(&req); err != nil {
			if err == io.EOF {
				return nil
			}
			_ = enc.Encode(s.errResp(nil, ErrParseError, "parse error"))
			// Drain the bad token so the decoder can continue.
			dec.Token() //nolint:errcheck
			continue
		}
		resp := s.handle(&req)
		if resp != nil {
			if err := enc.Encode(resp); err != nil {
				fmt.Fprintf(os.Stderr, "encode error: %v\n", err)
			}
		}
	}
}

func (s *Server) handle(req *Request) *Response {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "notifications/initialized":
		return nil // notification — no response
	case "tools/list":
		return s.handleListTools(req)
	case "tools/call":
		return s.handleCallTool(req)
	case "ping":
		return s.okResp(req.ID, map[string]interface{}{})
	default:
		return s.errResp(req.ID, ErrMethodNotFound, "method not found: "+req.Method)
	}
}

func (s *Server) handleInitialize(req *Request) *Response {
	return s.okResp(req.ID, InitializeResult{
		ProtocolVersion: ProtocolVersion,
		ServerInfo:      ServerInfo{Name: "tradingview", Version: serverVersion},
		Capabilities:    map[string]interface{}{"tools": map[string]interface{}{}},
		Instructions:    s.instructions,
	})
}

func (s *Server) handleListTools(req *Request) *Response {
	return s.okResp(req.ID, ListToolsResult{Tools: s.registry.List()})
}

func (s *Server) handleCallTool(req *Request) *Response {
	var params CallToolParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return s.errResp(req.ID, ErrInvalidParams, "invalid params")
	}
	result, err := s.registry.Call(params.Name, params.Arguments)
	if err != nil {
		return s.okResp(req.ID, CallToolResult{
			Content: []ContentItem{{Type: "text", Text: err.Error()}},
			IsError: true,
		})
	}
	text, err := json.Marshal(result)
	if err != nil {
		return s.errResp(req.ID, ErrInternal, "marshal error")
	}
	return s.okResp(req.ID, CallToolResult{
		Content: []ContentItem{{Type: "text", Text: string(text)}},
	})
}

func (s *Server) okResp(id json.RawMessage, result interface{}) *Response {
	return &Response{JSONRPC: "2.0", ID: id, Result: result}
}

func (s *Server) errResp(id json.RawMessage, code int, msg string) *Response {
	return &Response{JSONRPC: "2.0", ID: id, Error: &RPCError{Code: code, Message: msg}}
}
