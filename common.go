package netstring

import (
	"errors"
)

// MaximumLength defines the maximum length of a value in a netstring.
//
// The original specification doesn't actually define a maximum length so this somewhat
// arbitrary value is defined mostly as a safety margin for CPUs for which the go compiler
// defines int as int32.
//
// Having said that, the original specification *does* include a code fragment which
// suggests the same limit so it seems like a good place to start. This limit is slighty
// less than 2^30, so safe for any int32/uint32 storage.
const MaximumLength = 999999999

const (
	leadingColon  byte = ':'
	trailingComma byte = ','

	errorPrefix = "netstring: "
)

var (
	trueByte  = []byte{'T'}
	falseByte = []byte{'f'}

	leadingDelimiter  = []byte{leadingColon}
	trailingDelimiter = []byte{trailingComma}
)

var ErrLengthNotDigit = errors.New(errorPrefix + "Length does not start with a digit")
var ErrLeadingZero = errors.New(errorPrefix + "Non-zero length cannot have a leading zero")
var ErrLengthToLong = errors.New(errorPrefix + "Length contains more bytes than maximum allowed")
var ErrValueToLong = errors.New(errorPrefix + "Length of value is longer than maximum allowed")
var ErrColonExpected = errors.New(errorPrefix + "Leading colon delimiter not found after length")
var ErrCommaExpected = errors.New(errorPrefix + "Trailing comma delimeter not found after value")

var ErrNoKey = errors.New(errorPrefix + "Keyed netstring cannot be NoKey")
var ErrUnsupportedType = errors.New(errorPrefix + "Unsupported go type supplied to Encode()")
var ErrZeroKey = errors.New(errorPrefix + "Keyed netstring is zero length (thus has no key)")
var ErrInvalidKey = errors.New(errorPrefix + "Key is not in range 'a'-'z' or 'A'-'Z'")

var ErrBadMarshalValue = errors.New(errorPrefix + "Marshal only accepts struct{} and *struct{}")
var ErrBadMarshalTag = errors.New(errorPrefix + "struct tag is not a valid netstring.Key")
var ErrBadUnmarshalMsg = errors.New(errorPrefix + "Unmarshal only accepts *struct{}")
var ErrBadMarshalEOM = errors.New(errorPrefix + "End-of-Message Key is invalid")
