// Package mcp — error classification for HTS consumers.
package mcp

import "strings"

// retryableSubstrings are substrings whose presence in an error message
// indicates a transient condition that may resolve on retry.
var retryableSubstrings = []string{
	"CDP",
	"connect",
	"no TradingView",
	"timeout",
	"websocket",
	"WebSocket",
}

// permanentSubstrings indicate errors that will not resolve on retry.
var permanentSubstrings = []string{
	"unknown tool",
	"unmarshal",
	"invalid",
	"entity_id is required",
	"name is required",
	"query is required",
}

// IsRetryable reports whether err is a transient error worth retrying.
// Returns false for nil errors.
//
// Transient: CDP disconnect, TradingView tab not found, JS timeout.
// Permanent: unknown tool name, bad argument types, missing required fields.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	for _, sub := range permanentSubstrings {
		if strings.Contains(s, sub) {
			return false
		}
	}
	for _, sub := range retryableSubstrings {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// ErrorKind categorises an error for structured handling by consumers.
type ErrorKind string

const (
	ErrKindCDPDisconnect  ErrorKind = "cdp_disconnect"  // CDP not reachable — retryable
	ErrKindTabNotFound    ErrorKind = "tab_not_found"    // TradingView tab absent — retryable
	ErrKindJSTimeout      ErrorKind = "js_timeout"       // evaluation timed out — retryable
	ErrKindUnknownTool    ErrorKind = "unknown_tool"     // no such MCP tool — permanent
	ErrKindBadArgs        ErrorKind = "bad_args"         // argument parse/validation — permanent
	ErrKindUnknown        ErrorKind = "unknown"          // uncategorised
)

// ClassifyError returns the ErrorKind for a given error.
func ClassifyError(err error) ErrorKind {
	if err == nil {
		return ErrKindUnknown
	}
	s := err.Error()
	switch {
	case strings.Contains(s, "unknown tool"):
		return ErrKindUnknownTool
	case strings.Contains(s, "unmarshal") || strings.Contains(s, "invalid") ||
		strings.Contains(s, "is required"):
		return ErrKindBadArgs
	case strings.Contains(s, "timeout"):
		return ErrKindJSTimeout
	case strings.Contains(s, "no TradingView"):
		return ErrKindTabNotFound
	case strings.Contains(s, "CDP") || strings.Contains(s, "connect") ||
		strings.Contains(s, "websocket") || strings.Contains(s, "WebSocket"):
		return ErrKindCDPDisconnect
	default:
		return ErrKindUnknown
	}
}
