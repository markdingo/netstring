package netstring_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/markdingo/netstring"
)

func TestMarshal(t *testing.T) {
	type structA struct {
		Age         int    `netstring:"a"`
		Country     string `netstring:"c"`
		TLD         []byte `netstring:"t"`
		CountryCode []byte `netstring:"C"`
		Name        string `netstring:"n"`
		Height      uint16 // Ignored - no netstring tag
		dbKey       int64  // Ignored - not exported
	}

	type structB struct {
		A int32   `netstring:"A"`
		B int64   `netstring:"B"`
		C uint32  `netstring:"a"`
		D uint64  `netstring:"b"`
		E float32 `netstring:"C"`
		F float64 `netstring:"c"`
	}

	type structC struct {
		A int32   `netstring:"A"`
		B int64   `netstring:"B"`
		C uint32  `netstring:"A"` // Duplicate tag
		D uint64  `netstring:"b"`
		E float32 `netstring:"C"`
		F float64 `netstring:"c"`
	}

	type structD struct {
		Ad int32    `netstring:"A"`
		Bd int64    `netstring:"B"`
		Cd []string `netstring:"C"` // Not a simple struct
		Dd uint64   `netstring:"D"`
		Ed float32  `netstring:"E"`
	}

	type structE struct {
		A int32       `netstring:"A"`
		B int64       `netstring:"B"`
		C map[int]int `netstring:"C"` // Not a simple struct
		D uint64      `netstring:"D"`
		E float32     `netstring:"E"`
	}

	type structF struct {
		A int32 `netstring:"AA"` // Wrong tag length
	}

	type structG struct {
		A int32 `netstring:"$"` // Invalid Tag
	}

	type structH struct {
		Ah int32 `netstring:"\000"` // Not a keyed tag
	}

	type testCase struct {
		eom           netstring.Key
		message       any
		errorContains string
		expect        string
	}

	a1 := structA{21, "Iceland", []byte{'i', 'c'}, []byte("354"), "Bjorn", 183, 123456}
	b1 := structB{-20, -300, 21, 301, 123.5, -234.56}
	c1 := structC{-20, -300, 21, 301, 123.5, -234.56}
	d1 := structD{}
	e1 := structE{}
	f1 := structF{}
	g1 := structG{}
	h1 := structH{}

	testCases := []testCase{
		{'Z', nil, "Marshal only accepts", ""},                                 // 0
		{'Z', int(50), "Marshal only accepts", ""},                             // 1
		{'Z', a1, "", "3:a21,8:cIceland,3:tic,4:C354,6:nBjorn,1:Z,"},           // 2
		{'Z', &a1, "", "3:a21,8:cIceland,3:tic,4:C354,6:nBjorn,1:Z,"},          // 3 Pointer to should be same
		{'Z', &b1, "", "4:A-20,5:B-300,3:a21,4:b301,6:C123.5,8:c-234.56,1:Z,"}, // 4
		{'Z', c1, "Duplicate tag", ""},                                         // 5
		{'Z', d1, "unsupported", ""},                                           // 6
		{'Z', e1, "unsupported", ""},                                           // 7
		{'Z', f1, "not a valid", ""},                                           // 8
		{'Z', g1, "not in range", ""},                                          // 9
		{'Z', h1, "not a valid", ""},                                           // 10
		{netstring.NoKey, a1, "End-of-Message Key is invalid", ""},             // 11
		{'$', a1, "Key is not in range", ""},                                   // 12
	}

	for ix, tc := range testCases {
		var bbuf bytes.Buffer
		enc := netstring.NewEncoder(&bbuf)
		err := enc.Marshal(tc.eom, tc.message)
		if err != nil {
			if len(tc.errorContains) == 0 {
				t.Error(ix, "Unexpected", err.Error())
				continue
			}
			if !strings.Contains(err.Error(), tc.errorContains) {
				t.Error(ix, "Wrong Error", err.Error())
			}
			continue // Got expected error
		} else {
			if len(tc.errorContains) > 0 {
				t.Error(ix, "Expected error with", tc.errorContains)
				continue
			}
		}

		actual := bbuf.String()
		if actual != tc.expect {
			t.Error(ix, "Wrong result\nGot", actual, "\nExp", tc.expect)
		}
	}
}
