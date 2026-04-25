package pane

import (
	"testing"

	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
)

func TestResolveLayoutKnownCodes(t *testing.T) {
	cases := []struct{ in, want string }{
		{"s", "s"},
		{"2h", "2h"},
		{"4", "4"},
		{"16", "16"},
	}
	for _, tc := range cases {
		got, err := resolveLayout(tc.in)
		if err != nil {
			t.Errorf("resolveLayout(%q) error: %v", tc.in, err)
		}
		if got != tc.want {
			t.Errorf("resolveLayout(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestResolveLayoutAliases(t *testing.T) {
	cases := []struct{ in, want string }{
		{"single", "s"},
		{"1", "s"},
		{"1x1", "s"},
		{"2x2", "4"},
		{"quad", "4"},
		{"grid", "4"},
		{"2x1", "2h"},
		{"1x2", "2v"},
	}
	for _, tc := range cases {
		got, err := resolveLayout(tc.in)
		if err != nil {
			t.Errorf("resolveLayout(%q) error: %v", tc.in, err)
		}
		if got != tc.want {
			t.Errorf("resolveLayout(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestResolveLayoutCaseInsensitive(t *testing.T) {
	got, err := resolveLayout("SINGLE")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "s" {
		t.Errorf("got %q, want s", got)
	}
}

func TestResolveLayoutUnknown(t *testing.T) {
	_, err := resolveLayout("bogus")
	if err == nil {
		t.Error("expected error for unknown layout")
	}
}

func TestRegisterPaneToolNames(t *testing.T) {
	reg := mcp.NewRegistry()
	RegisterTools(reg)
	tools := reg.List()
	names := make(map[string]bool, len(tools))
	for _, td := range tools {
		names[td.Name] = true
	}
	for _, want := range []string{"pane_list", "pane_set_layout", "pane_focus", "pane_set_symbol"} {
		if !names[want] {
			t.Errorf("tool %q not registered", want)
		}
	}
}
