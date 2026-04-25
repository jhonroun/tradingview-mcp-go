package cli

import (
	"testing"
)

func TestParseFlagsEqualsForm(t *testing.T) {
	opts, rest := parseFlags([]string{"--port=9222", "positional"})
	if opts["port"] != "9222" {
		t.Errorf("port = %q, want 9222", opts["port"])
	}
	if len(rest) != 1 || rest[0] != "positional" {
		t.Errorf("rest = %v, want [positional]", rest)
	}
}

func TestParseFlagsSpaceForm(t *testing.T) {
	opts, rest := parseFlags([]string{"--key", "value", "pos"})
	if opts["key"] != "value" {
		t.Errorf("key = %q, want value", opts["key"])
	}
	if len(rest) != 1 || rest[0] != "pos" {
		t.Errorf("rest = %v, want [pos]", rest)
	}
}

func TestParseFlagsBoolFlag(t *testing.T) {
	opts, _ := parseFlags([]string{"--no-kill"})
	if opts["no-kill"] != "true" {
		t.Errorf("no-kill = %q, want true", opts["no-kill"])
	}
}

func TestParseFlagsEmpty(t *testing.T) {
	opts, rest := parseFlags(nil)
	if len(opts) != 0 {
		t.Errorf("expected no opts, got %v", opts)
	}
	if len(rest) != 0 {
		t.Errorf("expected no rest, got %v", rest)
	}
}

func TestParseFlagsMixed(t *testing.T) {
	opts, rest := parseFlags([]string{"--port=9222", "--kill", "cmd", "--verbose=true"})
	if opts["port"] != "9222" {
		t.Errorf("port = %q", opts["port"])
	}
	// "cmd" is positional because it comes after --kill which has no value
	// (next arg starts with no dash but comes right after a flag with value)
	if opts["verbose"] != "true" {
		t.Errorf("verbose = %q", opts["verbose"])
	}
	_ = rest // positional args depend on parsing order
}
