package netstring

import (
	"fmt"
	"io"
	"strconv"
)

/*
Encoder provides Encode*() functions to encode basic go types as netstrings and write them
to the io.Writer. Encode also provides Marshal() which assembles a complete message from a
simple struct as a series of netstrings. An Encoder *must* be constructed with
NewEncoder() otherwise subsequent calls will panic.

The first parameter to every Encode*() function is a Key type called "key" which can be
either a binary '0' (aka netstring.NoKey) which causes the Encoder to emit a regular
netstring or any isalpha() value which causes the Encoder to emit a "keyed" netstring.
Any "key" value outside those ranges is invalid and results in an error return. The "key"
is tested using netstring.Key.Assess().

The "key" in "keyed" netstrings can be used to categorized the netstring in some
meaningful way for the application. In this case the receiving application calls
Decode.DecodeKeyed() to return this "key" and the rest of the netstring as a value.

"Keyed" netstrings simply mean that the "key" byte is the first byte of the netstring with
the value, if any, being the following bytes. It's nothing particularly fancy, but it does
afford the application signifcantly more flexibility as described in the general package
documentation.

Idiomatic use of Encoder is to supply a network socket to NewEncoder() thus encoded
netstrings are automatically written to the network. Similarly the receiver connects their
network socket to a Decoder() and automatically receive decoded netstrings as they arrive.

Almost all error returns will be errors from the underlying io.Writer which tends to mean
a Write() to a network socket failed.
*/
type Encoder struct {
	formatBuffer [40]byte // Easily fits MaximumLength bytes (and 2^64 as well)
	out          io.Writer
}

// NewEncoder constructs a netstring encoder. An Encoder *must* be constructed with
// NewEncoder otherwise subsequent calls will panic.
//
// Each call to a Encode*() function results in a netstring being written to the
// io.Writer, quite possibly with multiple Write() calls.
func NewEncoder(output io.Writer) *Encoder {
	return &Encoder{out: output}
}

// EncodeBytes encodes the variadic arguments as a series of bytes in a single netstring.
//
// This function returns an error if key.Assess() returns an error. If key ==
// netstring.NoKey then a standard netstring is encoded otherwise a "keyed" netstring is
// encoded.
//
// EncodeBytes is the recommended function to create an end-of-message sentinel for
// "keyed" netstring. If, e.g., a "key" of 'z' is the sentinel, then:
//
//	EncodeBytes('z')
//
// generates the appropriate "keyed" netstring.
func (enc *Encoder) EncodeBytes(key Key, val ...[]byte) error {
	var l uint64 // Calculate the length of the netstring
	keyed, err := key.Assess()
	if err != nil {
		return err
	}
	if keyed {
		l++
	}
	for _, subVal := range val {
		l += uint64(len(subVal))
	}
	if l > MaximumLength {
		return ErrValueToLong
	}

	// Write the decimal length of the value (via formatBuffer for performance reasons)
	ls := enc.formatBuffer[0:0:len(enc.formatBuffer)]
	ls = strconv.AppendUint(ls, l, 10)
	_, err = enc.out.Write(ls)
	if err != nil {
		return fmt.Errorf(errorPrefix+"Encoder write length failed: %w", err)
	}

	// Write the leading delimiter
	_, err = enc.out.Write(leadingDelimiter)
	if err != nil {
		return fmt.Errorf(errorPrefix+"Encoder write leading delimiter failed: %w", err)
	}

	// Write key if its "keyed"
	if keyed {
		// Write key (via formatBuffer to avoid allocation)
		enc.formatBuffer[0] = byte(key)
		_, err = enc.out.Write(enc.formatBuffer[0:1])
		if err != nil {
			return fmt.Errorf(errorPrefix+"Encoder write key failed: %w", err)
		}
	}

	// Write the values
	for _, subVal := range val {
		if len(subVal) > 0 {
			_, err = enc.out.Write(subVal)
			if err != nil {
				return fmt.Errorf(errorPrefix+"Encoder write value failed: %w", err)
			}
		}
	}

	// And finally write the trailing delimiter
	_, err = enc.out.Write(trailingDelimiter)
	if err != nil {
		return fmt.Errorf(errorPrefix+"Encoder write trailing delimiter failed: %w", err)
	}

	return nil
}

// EncodeString encodes a string as a netstring. If key == netstring.NoKey a standard
// netstring is encoded otherwise a "keyed" netstring is encoded. "key" must pass
// Key.Assess() otherwise an error is returned.
func (enc *Encoder) EncodeString(key Key, val string) error {
	return enc.EncodeBytes(key, []byte(val))
}

// EncodeBool encodes a boolean value as a netstring. If key == netstring.NoKey a standard
// netstring is encoded otherwise a "keyed" netstring is encoded. "key" must pass
// Key.Assess() otherwise an error is returned.
//
// Accepted strconv shorthand of 'T' and 'f' represents true and false
// respectively. Recommended conversion back to boolean is via strconv.ParseBool()
func (enc *Encoder) EncodeBool(key Key, val bool) error {

	if val {
		return enc.EncodeBytes(key, trueByte)
	}
	return enc.EncodeBytes(key, falseByte)
}

// EncodeInt encodes an int as a netstring using strconv.FormatInt. Recommended conversion
// back to int is via strconv.ParseInt(). "key" must pass Key.Assess() otherwise an error
// is returned.
func (enc *Encoder) EncodeInt(key Key, val int) error {
	return enc.EncodeString(key, strconv.FormatInt(int64(val), 10))
}

// EncodeInt encodes a uint as a netstring using strconv.FormatUint. Recommended
// conversion back to int is via strconv.ParseUint(). "key" must pass Key.Assess()
// otherwise an error is returned.
func (enc *Encoder) EncodeUint(key Key, val uint) error {
	return enc.EncodeString(key, strconv.FormatUint(uint64(val), 10))
}

// EncodeInt32 encodes an int32 as a netstring using strconv.FormatInt. "key" must pass
// Key.Assess() otherwise an error is returned.
func (enc *Encoder) EncodeInt32(key Key, val int32) error {
	return enc.EncodeString(key, strconv.FormatInt(int64(val), 10))
}

// EncodeUint32 encodes a uint32 as a netstring using strconv.FormatUInt. Recommended
// conversion back to int32 is via strconv.ParseInt(). "key" must pass Key.Assess()
// otherwise an error is returned.
func (enc *Encoder) EncodeUint32(key Key, val uint32) error {
	return enc.EncodeString(key, strconv.FormatUint(uint64(val), 10))
}

// EncodeInt64 encodes an int64 as a netstring using strconv.FormatInt. Recommended
// conversion back to int64 is via strconv.ParseInt(). "key" must pass Key.Assess()
// otherwise an error is returned.
func (enc *Encoder) EncodeInt64(key Key, val int64) error {
	return enc.EncodeString(key, strconv.FormatInt(val, 10))
}

// EncodeUint64 encodes a uint64 as a netstring using strconv.FormatUint. Recommended
// conversion back to int64 is via strconv.ParseUint(). "key" must pass Key.Assess()
// otherwise an error is returned.
func (enc *Encoder) EncodeUint64(key Key, val uint64) error {
	return enc.EncodeString(key, strconv.FormatUint(val, 10))
}

// EncodeFloat32 encodes a float32 as a netstring using strconv.FormatFloat with the 'f'
// format. Recommended conversion back to float32 is via strconv.ParseFloat(). "key" must
// pass Key.Assess() otherwise an error is returned.
func (enc *Encoder) EncodeFloat32(key Key, val float32) error {
	return enc.EncodeString(key, strconv.FormatFloat(float64(val), 'f', -1, 32))
}

// EncodeFloat64 encodes a float64 as a netstring using strconv.FormatFloat with the 'f'
// format. Recommended conversion back to float64 is via strconv.ParseFloat(). "key" must
// pass Key.Assess() otherwise an error is returned.
func (enc *Encoder) EncodeFloat64(key Key, val float64) error {
	return enc.EncodeString(key, strconv.FormatFloat(val, 'f', -1, 64))
}

// EncodeByte encodes a single byte as a netstring. "key" must pass Key.Assess() otherwise
// an error is returned.
func (enc *Encoder) EncodeByte(key Key, val byte) error {
	return enc.EncodeBytes(key, []byte{val})
}

// Encode is the type-generic function which encodes most simple go types. Encode() uses
// go type-casting of val.(type) to determine the type-specific encoder to call. "key"
// must pass Key.Assess() otherwise an error is returned.
//
// Be wary of encoding a rune (a single quoted unicode character) with Encode() as the go
// compiler arranges for a rune to be passed in as an int32 and will thus be encoded as a
// string representation of its integer value. Recipient applications need to be aware of
// this conversion if they want to reconstruct the original rune.
//
// A better strategy is to pass unicode characters to Encode() as a string and single
// bytes should be cast as a byte, e.g. Encode(0, byte('Z')). When in doubt it's best to
// use type-specific functions such as EncodeByte() and EncodeString().
func (enc *Encoder) Encode(key Key, val any) error {
	switch tval := val.(type) {
	case byte:
		return enc.EncodeByte(key, tval)
	case []byte:
		return enc.EncodeBytes(key, tval)
	case string:
		return enc.EncodeString(key, tval)
	case bool:
		return enc.EncodeBool(key, tval)
	case int:
		return enc.EncodeInt(key, tval)
	case uint:
		return enc.EncodeUint(key, uint(tval))
	case int32:
		return enc.EncodeInt32(key, int32(tval))
	case uint32:
		return enc.EncodeUint32(key, uint32(tval))
	case int64:
		return enc.EncodeInt64(key, int64(tval))
	case uint64:
		return enc.EncodeUint64(key, uint64(tval))
	case float32:
		return enc.EncodeFloat32(key, tval)
	case float64:
		return enc.EncodeFloat64(key, tval)
	}

	return ErrUnsupportedType
}
