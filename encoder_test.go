package netstring_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/markdingo/netstring"
)

func TestEncoderTypeSpecific(t *testing.T) {
	var bbuf bytes.Buffer
	e := netstring.NewEncoder(&bbuf)

	err := e.EncodeBytes(netstring.NoKey, []byte{'a', 'b', 'c'}, []byte{'x', 'y', 'z'})
	if err != nil {
		t.Fatal(err)
	}

	act := bbuf.String()
	exp := "6:abcxyz,"
	if act != exp {
		t.Error("EncodeBytes returned", act, "Expected", exp)
	}

	bbuf.Reset()
	e.EncodeBytes(netstring.NoKey, []byte{'A', 'B'})
	exp = "2:AB,"

	e.EncodeString(netstring.NoKey, "CD")
	exp += "2:CD,"

	e.EncodeBool(netstring.NoKey, true)
	exp += "1:T,"

	e.EncodeBool(netstring.NoKey, false)
	exp += "1:f,"

	e.EncodeInt(netstring.NoKey, -12345)
	exp += "6:-12345,"

	e.EncodeUint(netstring.NoKey, 678)
	exp += "3:678,"

	e.EncodeInt32(netstring.NoKey, -2345)
	exp += "5:-2345,"

	e.EncodeUint32(netstring.NoKey, 78)
	exp += "2:78,"

	e.EncodeInt64(netstring.NoKey, -234567890)
	exp += "10:-234567890,"

	e.EncodeUint64(netstring.NoKey, 7890123456789)
	exp += "13:7890123456789,"

	e.EncodeFloat32(netstring.NoKey, 12.34567)
	exp += "8:12.34567,"

	e.EncodeFloat64(netstring.NoKey, 12.3456789012345)
	exp += "16:12.3456789012345,"

	e.EncodeByte(netstring.NoKey, 'Z')
	exp += "1:Z,"

	act = bbuf.String()
	if act != exp {
		t.Error("Encode Types returned", act, "Expected", exp)
	}
}

func TestEncoderGeneric(t *testing.T) {
	var bbuf bytes.Buffer
	e := netstring.NewEncoder(&bbuf)

	err := e.Encode(0, []byte{'A', 'B'})
	if err != nil {
		t.Fatal(err)
	}
	exp := "2:AB,"

	err = e.Encode(0, "CD")
	if err != nil {
		t.Fatal(err)
	}
	exp += "2:CD,"

	err = e.Encode(0, true)
	if err != nil {
		t.Fatal(err)
	}
	exp += "1:T,"

	err = e.Encode(0, false)
	if err != nil {
		t.Fatal(err)
	}
	exp += "1:f,"

	err = e.Encode(0, int(-12345))
	if err != nil {
		t.Fatal(err)
	}
	exp += "6:-12345,"

	err = e.Encode(0, uint(678))
	if err != nil {
		t.Fatal(err)
	}
	exp += "3:678,"

	err = e.Encode(0, int32(-2345))
	if err != nil {
		t.Fatal(err)
	}
	exp += "5:-2345,"

	err = e.Encode(0, uint32(78))
	if err != nil {
		t.Fatal(err)
	}
	exp += "2:78,"

	err = e.Encode(0, int64(-234567890))
	if err != nil {
		t.Fatal(err)
	}
	exp += "10:-234567890,"

	err = e.Encode(0, uint64(7890123456789))
	if err != nil {
		t.Fatal(err)
	}
	exp += "13:7890123456789,"

	err = e.Encode(0, float32(12.34567))
	if err != nil {
		t.Fatal(err)
	}
	exp += "8:12.34567,"

	err = e.Encode(0, float64(12.3456789012345))
	if err != nil {
		t.Fatal(err)
	}
	exp += "16:12.3456789012345,"

	err = e.Encode(0, []byte{'Z'})
	if err != nil {
		t.Fatal(err)
	}
	exp += "1:Z,"

	err = e.EncodeBytes('z') // A zero-length keyed sentinel
	if err != nil {
		t.Fatal(err)
	}
	exp += "1:z,"

	act := bbuf.String()
	if act != exp {
		t.Error("Encode Types returned", act, "Expected", exp)
	}
}

func TestEncoderNoKey(t *testing.T) {
	var bbuf bytes.Buffer
	e := netstring.NewEncoder(&bbuf)

	err := e.Encode(netstring.NoKey, []byte{'A', 'B'})
	if err != nil {
		t.Fatal(err)
	}
	exp := "2:AB,"

	err = e.Encode(netstring.NoKey, "CD")
	if err != nil {
		t.Fatal(err)
	}
	exp += "2:CD,"

	err = e.Encode(netstring.NoKey, true)
	if err != nil {
		t.Fatal(err)
	}
	exp += "1:T,"

	err = e.Encode(netstring.NoKey, false)
	if err != nil {
		t.Fatal(err)
	}
	exp += "1:f,"

	err = e.EncodeInt(netstring.NoKey, int(-12345))
	if err != nil {
		t.Fatal(err)
	}
	exp += "6:-12345,"

	err = e.Encode(netstring.NoKey, uint(678))
	if err != nil {
		t.Fatal(err)
	}
	exp += "3:678,"

	err = e.Encode(netstring.NoKey, int32(-2345))
	if err != nil {
		t.Fatal(err)
	}
	exp += "5:-2345,"

	err = e.Encode(netstring.NoKey, uint32(78))
	if err != nil {
		t.Fatal(err)
	}
	exp += "2:78,"

	err = e.Encode(netstring.NoKey, int64(-234567890))
	if err != nil {
		t.Fatal(err)
	}
	exp += "10:-234567890,"

	err = e.Encode(netstring.NoKey, uint64(7890123456789))
	if err != nil {
		t.Fatal(err)
	}
	exp += "13:7890123456789,"

	err = e.Encode(netstring.NoKey, float32(12.34567))
	if err != nil {
		t.Fatal(err)
	}
	exp += "8:12.34567,"

	err = e.Encode(netstring.NoKey, float64(12.3456789012345))
	if err != nil {
		t.Fatal(err)
	}
	exp += "16:12.3456789012345,"

	err = e.Encode(netstring.NoKey, []byte{'Z'})
	if err != nil {
		t.Fatal(err)
	}
	exp += "1:Z,"

	act := bbuf.String()
	if act != exp {
		t.Error("Encode Types returned", act, "Expected", exp)
	}
}

func TestEncoderRune(t *testing.T) {
	var b bytes.Buffer
	e := netstring.NewEncoder(&b)

	err := e.Encode(0, 'A') // Rune turns into an int32
	if err != nil {
		t.Fatal(err)
	}
	exp := "2:65,"

	err = e.Encode(0, byte('A')) // Remains a byte
	if err != nil {
		t.Fatal(err)
	}
	exp += "1:A,"

	err = e.EncodeString(0, "®©") // Unicode Registered trademark, Apple, and Copyright symbols.
	if err != nil {
		t.Fatal(err)
	}
	exp += "7:®©,"

	act := b.String()
	if exp != act {
		t.Error("Type Coercion returned", act, "Expected", exp)
	}
}

func TestEncoderGenericBad(t *testing.T) {
	type someStruct struct {
		something     int
		somethingElse string
	}
	var b bytes.Buffer
	var s someStruct
	e := netstring.NewEncoder(&b)
	err := e.Encode(netstring.NoKey, s)
	if err == nil {
		t.Error("Expected error when trying to encode generic struct")
	}
}

type badWriter struct {
	when int
	err  string
}

func (bw *badWriter) Write(b []byte) (n int, err error) {
	if len(bw.err) > 0 {
		bw.when--
		if bw.when == 0 {
			return 0, errors.New(bw.err)
		}
	}

	return len(b), nil
}

func TestEncoderErrors(t *testing.T) {
	var bw badWriter
	e := netstring.NewEncoder(&bw)

	// Trigger write len error in EncodeBytes()

	bw.err = "WLen"
	bw.when = 1
	err := e.EncodeBytes('A', []byte{'A'})
	if err == nil {
		t.Fatal("Expected error return")
	}
	exp := "Encoder write length failed"
	if !strings.Contains(err.Error(), exp) {
		t.Error("Expected", exp, "in", err.Error())
	}

	// Trigger write leading delimiter error in EncodeBytes()

	bw.err = "WColon"
	bw.when = 2
	err = e.EncodeBytes('A', []byte{'A'})
	if err == nil {
		t.Fatal("Expected error return")
	}
	exp = "Encoder write leading delimiter failed"
	if !strings.Contains(err.Error(), exp) {
		t.Error("Expected", exp, "in", err.Error())
	}

	// Trigger write key error in EncodeBytes()

	bw.err = "WKey"
	bw.when = 3
	err = e.EncodeBytes('A', []byte{'A'})
	if err == nil {
		t.Fatal("Expected error return")
	}
	exp = "Encoder write key failed"
	if !strings.Contains(err.Error(), exp) {
		t.Error("Expected", exp, "in", err.Error())
	}

	// Trigger write value error in EncodeBytes()

	bw.err = "WValue"
	bw.when = 4
	err = e.EncodeBytes('A', []byte{'A'})
	if err == nil {
		t.Fatal("Expected error return")
	}
	exp = "Encoder write value failed"
	if !strings.Contains(err.Error(), exp) {
		t.Error("Expected", exp, "in", err.Error())
	}

	// Trigger write terminator error in EncodeBytes()

	bw.err = "WTerminator"
	bw.when = 5
	err = e.EncodeBytes('A', []byte{'A'})
	if err == nil {
		t.Fatal("Expected error return")
	}
	exp = "Encoder write trailing delimiter failed"
	if !strings.Contains(err.Error(), exp) {
		t.Error("Expected", exp, "in", err.Error())
	}
}

func TestEncoderInvalidKey(t *testing.T) {
	var b bytes.Buffer
	e := netstring.NewEncoder(&b)

	for _, k := range []netstring.Key{'a' - 1, 'z' + 1, 'A' - 1, 'Z' + 1, '0', '9'} {
		err := e.EncodeByte(k, 'A')
		if err != netstring.ErrInvalidKey {
			t.Error("Key of", string(k), "did not return ErrInvalidKey", err)
		}
	}
}
