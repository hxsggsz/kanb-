// Package semver provides minimal parsing and comparison for vMAJOR.MINOR.PATCH
// version strings, matching the tag format used by kanba's GitHub releases.
package semver

import (
	"fmt"
	"strconv"
	"strings"
)

type Version struct {
	Major, Minor, Patch int
}

func Parse(s string) (Version, error) {
	s = strings.TrimPrefix(s, "v")
	parts := strings.SplitN(s, ".", 3)
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version %q", s)
	}

	var nums [3]int
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return Version{}, fmt.Errorf("invalid version %q: %w", s, err)
		}
		nums[i] = n
	}

	return Version{Major: nums[0], Minor: nums[1], Patch: nums[2]}, nil
}

// Compare parses a and b and returns -1, 0, or 1 depending on whether a is
// less than, equal to, or greater than b.
func Compare(a, b string) (int, error) {
	va, err := Parse(a)
	if err != nil {
		return 0, err
	}
	vb, err := Parse(b)
	if err != nil {
		return 0, err
	}

	if va.Major != vb.Major {
		return cmpInt(va.Major, vb.Major), nil
	}
	if va.Minor != vb.Minor {
		return cmpInt(va.Minor, vb.Minor), nil
	}
	return cmpInt(va.Patch, vb.Patch), nil
}

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
