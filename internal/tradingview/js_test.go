package tradingview

import "testing"

func TestSafeString(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{`hello`, `"hello"`},
		{`say "hi"`, `"say \"hi\""`},
		{"tab\there", `"tab\there"`},
		{`back\slash`, `"back\\slash"`},
		{"line\nnewline", `"line\nnewline"`},
	}
	for _, c := range cases {
		got := SafeString(c.in)
		if got != c.want {
			t.Errorf("SafeString(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSafeStringNoInjection(t *testing.T) {
	// Backtick, template literal, and closing quote must all be escaped.
	dangerous := []string{
		"`injected`",
		"'); DROP TABLE--",
		`${alert(1)}`,
	}
	for _, s := range dangerous {
		got := SafeString(s)
		// Result must start and end with a double-quote (JSON string).
		if len(got) < 2 || got[0] != '"' || got[len(got)-1] != '"' {
			t.Errorf("SafeString(%q) = %q, not a JSON string", s, got)
		}
	}
}
