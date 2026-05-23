package updater_test

import (
	"testing"

	"github.com/t0mer/go-certi/internal/updater"
)

func TestCompare(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"2026.5.0", "2026.5.0", 0},
		{"2026.5.0", "2026.5.1", -1},
		{"2026.5.1", "2026.5.0", 1},
		{"2026.5.0", "2026.10.0", -1}, // numeric, not lexical
		{"2026.10.0", "2026.5.0", 1},
		{"2025.12.5", "2026.1.0", -1},
		{"dev", "2026.5.0", -1},
		{"2026.5.0", "dev", 1},
		{"dev", "dev", 0},
		{"", "2026.5.0", -1},
		{"v2026.5.0", "2026.5.0", 0},
	}
	for _, c := range cases {
		got := updater.Compare(c.a, c.b)
		if got != c.want {
			t.Errorf("Compare(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}
