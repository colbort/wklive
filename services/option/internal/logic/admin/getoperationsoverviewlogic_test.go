package adminlogic

import "testing"

func TestNormalizeOperationsStaleSeconds(t *testing.T) {
	tests := []struct {
		name  string
		input int64
		want  int64
		ok    bool
	}{
		{name: "default", input: 0, want: 60, ok: true},
		{name: "minimum", input: 10, want: 10, ok: true},
		{name: "maximum", input: 300, want: 300, ok: true},
		{name: "below minimum", input: 9, ok: false},
		{name: "above maximum", input: 301, ok: false},
		{name: "negative", input: -1, ok: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := normalizeOperationsStaleSeconds(test.input)
			if got != test.want || ok != test.ok {
				t.Fatalf("normalize(%d)=(%d,%v), want=(%d,%v)",
					test.input, got, ok, test.want, test.ok)
			}
		})
	}
}
