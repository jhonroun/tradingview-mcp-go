package mcp

import (
	"errors"
	"testing"
)

func TestIsRetryableNil(t *testing.T) {
	if IsRetryable(nil) {
		t.Error("nil error must not be retryable")
	}
}

func TestIsRetryableCDP(t *testing.T) {
	cases := []struct {
		msg  string
		want bool
	}{
		{"CDP connection refused", true},
		{"dial tcp: connect: connection refused", true},
		{"no TradingView tab found", true},
		{"context deadline exceeded: timeout", true},
		{"WebSocket closed", true},
		{"websocket read: EOF", true},
		{"unknown tool: foo", false},
		{"unmarshal error", false},
		{"invalid argument", false},
		{"entity_id is required", false},
		{"name is required", false},
		{"unrelated error", false},
	}
	for _, tc := range cases {
		err := errors.New(tc.msg)
		if got := IsRetryable(err); got != tc.want {
			t.Errorf("IsRetryable(%q) = %v, want %v", tc.msg, got, tc.want)
		}
	}
}

func TestClassifyError(t *testing.T) {
	cases := []struct {
		msg  string
		want ErrorKind
	}{
		{"unknown tool: foo", ErrKindUnknownTool},
		{"unmarshal failed", ErrKindBadArgs},
		{"invalid argument type", ErrKindBadArgs},
		{"entity_id is required", ErrKindBadArgs},
		{"context deadline exceeded: timeout", ErrKindJSTimeout},
		{"no TradingView tab found", ErrKindTabNotFound},
		{"CDP connect failed", ErrKindCDPDisconnect},
		{"websocket read EOF", ErrKindCDPDisconnect},
		{"some other error", ErrKindUnknown},
	}
	for _, tc := range cases {
		err := errors.New(tc.msg)
		if got := ClassifyError(err); got != tc.want {
			t.Errorf("ClassifyError(%q) = %q, want %q", tc.msg, got, tc.want)
		}
	}
}

func TestClassifyErrorNil(t *testing.T) {
	if got := ClassifyError(nil); got != ErrKindUnknown {
		t.Errorf("ClassifyError(nil) = %q, want %q", got, ErrKindUnknown)
	}
}
