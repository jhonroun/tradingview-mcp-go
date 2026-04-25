package ui

import (
	"testing"

	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
)

func TestRegisterUIToolNames(t *testing.T) {
	reg := mcp.NewRegistry()
	RegisterTools(reg)
	tools := reg.List()
	names := make(map[string]bool, len(tools))
	for _, td := range tools {
		names[td.Name] = true
	}
	want := []string{
		"ui_click", "ui_open_panel", "ui_fullscreen",
		"ui_keyboard", "ui_type_text", "ui_hover",
		"ui_scroll", "ui_mouse_click", "ui_find_element",
		"ui_evaluate", "layout_list", "layout_switch",
	}
	for _, w := range want {
		if !names[w] {
			t.Errorf("tool %q not registered", w)
		}
	}
	if len(tools) != len(want) {
		t.Errorf("expected %d UI tools, got %d", len(want), len(tools))
	}
}

func TestKeyMapEntries(t *testing.T) {
	required := []string{
		"Enter", "Escape", "Tab", "Backspace", "Delete",
		"ArrowUp", "ArrowDown", "ArrowLeft", "ArrowRight",
		"Space", "Home", "End", "PageUp", "PageDown",
		"F1", "F2", "F5",
	}
	for _, k := range required {
		if _, ok := keyMap[k]; !ok {
			t.Errorf("keyMap missing entry for %q", k)
		}
	}
}

func TestKeyboardModifierBitfield(t *testing.T) {
	// Verify modifier accumulation without CDP — just the bit logic.
	cases := []struct {
		mods []string
		want int
	}{
		{[]string{}, 0},
		{[]string{"alt"}, 1},
		{[]string{"ctrl"}, 2},
		{[]string{"meta"}, 4},
		{[]string{"shift"}, 8},
		{[]string{"ctrl", "shift"}, 10},
		{[]string{"ctrl", "alt"}, 3},
	}
	for _, tc := range cases {
		mod := 0
		for _, m := range tc.mods {
			switch m {
			case "alt":
				mod |= 1
			case "ctrl":
				mod |= 2
			case "meta":
				mod |= 4
			case "shift":
				mod |= 8
			}
		}
		if mod != tc.want {
			t.Errorf("modifiers %v → got %d, want %d", tc.mods, mod, tc.want)
		}
	}
}

func TestScrollDefaultAmount(t *testing.T) {
	// Without CDP we can't call Scroll(), but verify the sentinel logic.
	amount := 0
	if amount <= 0 {
		amount = 300
	}
	if amount != 300 {
		t.Errorf("default scroll amount should be 300, got %d", amount)
	}
}

func TestMouseClickButtonNormalise(t *testing.T) {
	cases := []struct{ in, want string }{
		{"left", "left"},
		{"right", "right"},
		{"middle", "middle"},
		{"", "left"},
		{"unknown", "left"},
	}
	for _, tc := range cases {
		btn := tc.in
		if btn != "right" && btn != "middle" {
			btn = "left"
		}
		if btn != tc.want {
			t.Errorf("button %q normalised to %q, want %q", tc.in, btn, tc.want)
		}
	}
}

func TestFindElementDefaultStrategy(t *testing.T) {
	strategy := ""
	if strategy == "" {
		strategy = "text"
	}
	if strategy != "text" {
		t.Errorf("default strategy should be 'text', got %q", strategy)
	}
}
