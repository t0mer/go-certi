package updater

import (
	"strconv"
	"strings"
)

// Compare returns -1 if a < b, 0 if a == b, 1 if a > b.
// Handles YYYY.M.PATCH versions with numeric per-segment comparison
// (so 2026.10.0 > 2026.5.0, unlike string compare). "dev" / empty is
// treated as the oldest possible version so any release looks newer.
func Compare(a, b string) int {
	if a == b {
		return 0
	}
	aDev := a == "" || a == "dev"
	bDev := b == "" || b == "dev"
	switch {
	case aDev && bDev:
		return 0
	case aDev:
		return -1
	case bDev:
		return 1
	}
	ap := strings.Split(strings.TrimPrefix(a, "v"), ".")
	bp := strings.Split(strings.TrimPrefix(b, "v"), ".")
	n := len(ap)
	if len(bp) > n {
		n = len(bp)
	}
	for i := 0; i < n; i++ {
		var ai, bi int
		if i < len(ap) {
			ai, _ = strconv.Atoi(ap[i])
		}
		if i < len(bp) {
			bi, _ = strconv.Atoi(bp[i])
		}
		if ai != bi {
			if ai < bi {
				return -1
			}
			return 1
		}
	}
	return 0
}
