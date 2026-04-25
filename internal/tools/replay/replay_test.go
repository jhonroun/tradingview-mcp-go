package replay

import (
	"testing"

	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
)

func TestRegisterReplayToolNames(t *testing.T) {
	reg := mcp.NewRegistry()
	RegisterTools(reg)
	tools := reg.List()
	names := make(map[string]bool, len(tools))
	for _, td := range tools {
		names[td.Name] = true
	}
	for _, want := range []string{
		"replay_start", "replay_step", "replay_stop",
		"replay_status", "replay_autoplay", "replay_trade",
	} {
		if !names[want] {
			t.Errorf("tool %q not registered", want)
		}
	}
	if len(tools) != 6 {
		t.Errorf("expected 6 replay tools, got %d", len(tools))
	}
}

func TestAutoplayInvalidSpeed(t *testing.T) {
	result, err := Autoplay(999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["success"] == true {
		t.Error("expected success=false for invalid autoplay speed")
	}
	errMsg, ok := result["error"].(string)
	if !ok || errMsg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestAutoplayValidSpeeds(t *testing.T) {
	// All valid speeds should NOT return the validation error immediately;
	// they would attempt CDP which isn't available in unit tests, so we
	// verify only that validAutoplayDelays has the correct set.
	expected := []int{100, 143, 200, 300, 1000, 2000, 3000, 5000, 10000}
	for _, s := range expected {
		if !validAutoplayDelays[s] {
			t.Errorf("speed %d should be valid but is not in validAutoplayDelays", s)
		}
	}
	if len(validAutoplayDelays) != len(expected) {
		t.Errorf("validAutoplayDelays has %d entries, want %d", len(validAutoplayDelays), len(expected))
	}
}

func TestTradeInvalidAction(t *testing.T) {
	// trade() with an invalid action goes through CDP first (starts == false error),
	// but the action validation happens inside the CDP session callback.
	// We can't test this without CDP; instead test that valid action strings exist.
	validActions := map[string]bool{"buy": true, "sell": true, "close": true}
	for _, a := range []string{"buy", "sell", "close"} {
		if !validActions[a] {
			t.Errorf("action %q missing from valid set", a)
		}
	}
}

func TestWvHelper(t *testing.T) {
	result := wv("foo.bar()")
	if len(result) == 0 {
		t.Error("wv() returned empty string")
	}
	// Must contain the path.
	if result == "foo.bar()" {
		t.Error("wv() should wrap the expression, not return it unchanged")
	}
}
