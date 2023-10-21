package netstring_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/markdingo/netstring"
)

func TestUnmarshal(t *testing.T) {
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

	type structF struct {
		A int32 `netstring:"AA"` // Wrong tag length
	}

	type structG struct {
		A int32 `netstring:"$"` // Invalid Tag
	}

	type structH struct {
		Ah int32 `netstring:"\000"` // Not a keyed tag
	}

	type structI struct {
		AI [10]int `netstring:"I"` // Not a basic type
	}

	type structJ struct {
		AJ []int `netstring:"I"` // Not a basic type
	}

	type structK struct {
		AK string `netstring:"s"`
	}

	type structL struct {
		IntTooBig   int8    `netstring:"b"` // Too big for uint8
		Negative    uint16  `netstring:"n"` // Negative value in unsigned
		FloatTooBig float32 `netstring:"f"` // Too big for float32
	}

	type structM struct {
		M1 int `netstring:"a"` // Duplicate and unknown
	}

	type testCase struct {
		eom           netstring.Key
		input         string
		errorContains string
		message       any
		expect        any
		unknown       netstring.Key
	}

	aexp := structA{21, "Iceland", []byte{'i', 'c'}, []byte("354"), "Bjorn", 0, 0}
	bexp := structB{-20, -300, 21, 301, 123.5, -234.56}
	iin := &structI{}

	testCases := []testCase{
		{'Z', "", "Marshal only accepts", nil, nil, 0},                                          // 0
		{'Z', "", "only accepts", int(50), nil, 0},                                              // 1
		{'Z', "3:a21,8:cIceland,3:tic,4:C354,6:nBjorn,1:Z,", "", &structA{}, &aexp, 0},          // 2
		{'Z', "4:A-20,5:B-300,3:a21,4:b301,6:C123.5,8:c-234.56,1:Z,", "", &structB{}, &bexp, 0}, // 3
		{'Z', "", "Duplicate tag", &structC{}, nil, 0},                                          // 4
		{'Z', "", "not a single character", &structF{}, nil, 0},                                 // 5
		{'Z', "", "not in range", &structG{}, nil, 0},                                           // 6
		{'Z', "", "not a valid", &structH{}, nil, 0},                                            // 7
		{'Z', "", "type unsupported", iin, nil, 0},                                              // 8
		{'Z', "", "only accepts", &iin, nil, 0},                                                 // 9
		{'Z', "", "type unsupported", &structJ{}, nil, 0},                                       // 10
		{'Z', "", "EOF", &structK{}, nil, 0},                                                    // 11
		{'A', "5:b1234,1:A,", "to int", &structL{}, &structL{}, 0},                              // 12
		{'A', "6:n-1234,1:A,", "overflows uint16", &structL{}, &structL{}, 0},                   // 13
		{'A', "8:f3.5e+38,1:A,", "overflows float32", &structL{}, &structL{}, 0},                // 14
		{'A', "4:a123,5:a2345,1:A,", "Duplicate", &structM{}, &structM{}, 0},                    // 15
		{'A', "4:b123,1:A,", "", &structM{}, &structM{}, 'b'},                                   // 16
		{netstring.NoKey, "4:b123,1:A,", "Key is invalid", &structM{}, &structM{}, 'b'},         // 17
		{'$', "4:b123,1:A,", "Key is not in range", &structM{}, &structM{}, 'b'},                // 18
	}

	for ix, tc := range testCases {
		bbuf := bytes.NewBufferString(tc.input)
		dec := netstring.NewDecoder(bbuf)
		unknown, err := dec.Unmarshal(tc.eom, tc.message)
		if err != nil {
			if len(tc.errorContains) == 0 {
				t.Error(ix, "Unexpected", err.Error())
			} else if !strings.Contains(err.Error(), tc.errorContains) {
				t.Error(ix, "Wrong Error", err.Error())
			}
			continue // Got expected error
		} else {
			if len(tc.errorContains) > 0 {
				t.Error(ix, "Expected error with", tc.errorContains)
				continue
			}
		}
		if unknown != tc.unknown {
			t.Error(ix, "Unexpected 'unknown' returned '%s'", unknown.String())
			continue
		}
		if tc.expect != nil {
			if !reflect.DeepEqual(tc.expect, tc.message) {
				t.Error(ix, "Comparison failed. \nExp", tc.expect, "\nGot", tc.message)
			}
		}
	}
}
