package netstring_test

import (
	"testing"

	"github.com/markdingo/netstring"
)

func TestKeyAssess(t *testing.T) {
	type testCase struct {
		k            netstring.Key
		keyed, error bool
	}

	testCases := []testCase{
		{0, false, false},
		{netstring.NoKey, false, false},
		{'0', false, true},
		{'9', false, true},
		{'a', true, false},
		{'z', true, false},
		{'A', true, false},
		{'Z', true, false},
		{'a' - 1, false, true}, // The next four tests assume  a gap between lowercase and uppercase
		{'z' + 1, false, true}, // letters, which happens to be true for ASCII and EBCDIC. Are there any
		{'A' - 1, false, true}, // other character sets we need to worry about?
		{'Z' + 1, false, true},
	}

	for ix, tc := range testCases {
		keyed, err := tc.k.Assess()
		if tc.error {
			if err == nil {
				t.Errorf("%d: %s should have returned an error", ix, string(tc.k))
				continue // keyed is undefined in this case so don't check
			}
		} else {
			if err != nil {
				t.Errorf("%d: %s gave unexpected error %s", ix, string(tc.k), err.Error())
				continue // keyed is undefined in this case so don't check
			}
		}

		if keyed != tc.keyed {
			t.Errorf("%d: %s gives keyed return. Expected %t got %t", ix, string(tc.k), tc.keyed, keyed)
		}
	}
}
