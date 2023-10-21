package netstring_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/markdingo/netstring"
)

func newWith(s string) *netstring.Decoder {
	return netstring.NewDecoder(bytes.NewBufferString(s))
}

func TestSimpleDecoder(t *testing.T) {
	dc := newWith("3:abc,4:wxyz,")
	v, e := dc.Decode()
	if e != nil {
		t.Fatal("Unexpected error", e)
	}

	if string(v) != "abc" {
		t.Error("Expected 'abc' value, but got", len(v), string(v))
	}

	v, e = dc.Decode()
	if e != nil {
		t.Fatal("Unexpected error", e)
	}

	if string(v) != "wxyz" {
		t.Error("Expected 'wxyz' value, but got", len(v), string(v))
	}
}

func TestDecoderErrors(t *testing.T) {
	type testCase struct {
		input string
		err   error
	}
	testCases := []testCase{
		{":abc,1:A,", netstring.ErrLengthNotDigit},
		{"03:abc,1:A,", netstring.ErrLeadingZero},
		{"999999999999:abc,1:A,", netstring.ErrLengthToLong},
		{"3*abc,1:A,", netstring.ErrColonExpected},
		{"3:abcZ1:A,", netstring.ErrCommaExpected},
	}

	for ix, tc := range testCases {
		dc := newWith(tc.input)
		_, err := dc.Decode()
		if err == nil {
			t.Error(ix, "Expected error return from", tc.input)
			continue
		}
		if err != tc.err {
			t.Error(ix, "Wrong error returned", err)
		}

		_, err = dc.Decode() // Second and subsequent should error
		if err != tc.err {
			t.Error(ix, "Wrong error returned", err)
		}
	}
}

type myReader struct {
	ch  chan []byte
	buf []byte
}

func newMyReader() *myReader {
	return &myReader{ch: make(chan []byte, 100)}
}

func (mr *myReader) set(p []byte) {
	mr.ch <- p
}
func (mr *myReader) close() {
	close(mr.ch)
}

func (mr *myReader) Read(p []byte) (n int, err error) {
	if len(mr.buf) == 0 {
		var ok bool
		mr.buf, ok = <-mr.ch
		if !ok {
			return 0, io.EOF
		}
	}
	n = copy(p, mr.buf)
	mr.buf = mr.buf[n:]

	return
}

func TestDecoderBuffer(t *testing.T) {
	type ns struct {
		key   netstring.Key
		value string
	}
	type testCase struct {
		input  string
		expect []ns
	}
	testCases := []testCase{
		// Partials
		{"13:Xzeroefghzero,", []ns{{'X', "zeroefghzero"}}},

		{"13:Xonedefghione", []ns{}},
		{",", []ns{{'X', "onedefghione"}}},

		{"13:Xtwodefghi", []ns{}},
		{"two,", []ns{{'X', "twodefghitwo"}}},

		{"13:Xt", []ns{}},
		{"hreefgthree,", []ns{{'X', "threefgthree"}}},

		{"13:X", []ns{}},
		{"fourefghfour,", []ns{{'X', "fourefghfour"}}},

		{"13", []ns{}},
		{":Xfiveefghfive,", []ns{{'X', "fiveefghfive"}}},

		{"1", []ns{}},
		{"3:Xsixdefghisix,", []ns{{'X', "sixdefghisix"}}},

		// Multiples

		{"2:w1,3:x22,4:y333,5:z4444", []ns{{'w', "1"}, {'x', "22"}, {'y', "333"}}},
		{",6:T55555,", []ns{{'z', "4444"}, {'T', "55555"}}},
	}

	mr := newMyReader()
	dc := netstring.NewDecoder(mr)
	for _, tc := range testCases { // Populate pipeline for multiple reads
		mr.set([]byte(tc.input)) // to exerise decoder buffer management
	}
	mr.close()
	for ix, tc := range testCases {
		for nsx, nse := range tc.expect {
			k, v, e := dc.DecodeKeyed()
			if e != nil {
				t.Fatal(ix, nsx, e)
			}
			if k != nse.key {
				t.Error(ix, nsx, "Wrong key", string(k), "Expected", string(nse.key))
			}
			if string(v) != nse.value {
				t.Error(ix, nsx, "Wrong value", string(v), "Expected", string(nse.value))
			}
		}
	}
	v, err := dc.Decode()
	if v != nil {
		t.Error("Expected nil value after DecodeKey() depleted", string(v), err)
	}
}

func TestDecoderDecodeKeyedErrors(t *testing.T) {
	dc := newWith("0:,")
	_, _, e := dc.DecodeKeyed()
	if e != netstring.ErrZeroKey {
		t.Error("Expected ZeroKey error, not", e)
	}

	dc = newWith("2:@1,") // Not isalpha()
	_, _, e = dc.DecodeKeyed()
	if e != netstring.ErrInvalidKey {
		t.Error("Expected InvalidKey error, not", e)
	}

	dc = newWith(string([]byte{'1', ':', 0, ','})) // NoKey as a keyed value is invalid
	_, _, e = dc.DecodeKeyed()
	if e != netstring.ErrInvalidKey {
		t.Error("Expected InvalidKey error, not", e)
	}
}

// Test that Write returns a perpetual error once one has been created by the parser.
func TestDecoderPerpetualWriteError(t *testing.T) {
	dc := newWith("aa1:a,") // Invalid length
	_, err := dc.Decode()
	if err != netstring.ErrLengthNotDigit {
		t.Fatal("Wrong first error returned", err)
	}
	_, err = dc.Decode() // Good netstring is irrelevant now
	if err != netstring.ErrLengthNotDigit {
		t.Fatal("Wrong second error returned", err)
	}
}

// Test that Decode returns all valid netstrings upto the parsing error
func TestDecoderPerpetualError(t *testing.T) {
	dc := newWith("1:a,2:bb,03:ccc,")

	// Decode should return the first two netstrings then an error after that as "03:"
	// is an invalid length.

	val, err := dc.Decode()
	if err != nil {
		t.Fatal("Unexpected error on first Decode()", err)
	}
	if string(val) != "a" {
		t.Fatal("Unexpected value", string(val))
	}

	val, err = dc.Decode()
	if err != nil {
		t.Fatal("Unexpected error on second Decode()", err)
	}
	if string(val) != "bb" {
		t.Fatal("Unexpected value", string(val))
	}

	// Now we should get an error in perpetuity due to the leading '0' in 03:ccc,

	_, err = dc.Decode()
	if err != netstring.ErrLeadingZero {
		t.Fatal("Expected error return due to leading zero, not", err)
	}

	// Make sure it's not a once-off
	_, err = dc.Decode()
	if err != netstring.ErrLeadingZero {
		t.Fatal("Expected error return due to leading zero, not", err)
	}
}

func TestDecodeKeyedWithNil(t *testing.T) {
	dc := newWith("")
	k, v, e := dc.DecodeKeyed()
	if e != io.EOF {
		t.Error("Expected EOF from empty parse but got", k, v, e)
	}
}
