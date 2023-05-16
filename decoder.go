package netstring

import (
	"io"
)

// parseState represents the state transitions for parsing a netstring. Different
// variables in the Decoder are valid or in use depending on parseState.
type parseState int

const (
	parseFirstByte parseState = iota // length, lengthBytesSeen
	parseLength                      // length, lengthBytesSeen
	parseColon
	parseValue // ns.value
	parseComma
)

// Only used for debugging purposes
func (t parseState) String() string {
	switch t {
	case parseFirstByte:
		return "parseFirstByte"
	case parseLength:
		return "parseLength"
	case parseColon:
		return "parseColon"
	case parseValue:
		return "parseValue"
	case parseComma:
		return "parseComma"
	}

	return "Bizarre parseState"
}

/*
Decoder provides a netstring decode capability. A Decoder *must* be constructed with
NewDecoder() otherwise subsequent calls will panic.

The byte-stream from the io.Reader provided to NewDecoder() is expected to contain a pure
stream of netstrings. Each netstring can be retrieved via Decode() and DecodeKeyed() for
standard netstrings and "keyed" netstrings respectively. The sending and receiving
applications must agree on all aspects of how these netstrings are interpreted. Typically
they will agree on a message structure which is either a fixed number of standard
netstrings or a variable number of "keyed" netstrings terminated by an end-of-message
sentinel.

The functions Decode() and DecodeKeyed() are used to accessed each decoded netstring as it
becomes available and Unmarshal() is used to decoded a complete "message" containing a
series of "keyed" netstrings (including an end-of-message sentinel) into a simple struct.

It is often good practice to wrap the input io.Reader in a bufio.Reader as this will
likely improve parsing performance of this package.

If the Decoder detects a malformed netstring, it stops parsing, returns an error and
effective stops all future parsing for that byte stream because once synchronization is
lost, it can never be recovered.

Decoder passes thru io.EOF from the io.Reader, but only after all bytes have been consumed
in the process of producing netstrings. An application should anticipate io.EOF if the
io.Reader constitutes a network connection of some type. Unlike io.Reader, the EOF error
is *not* returned in the same call which returns a valid netstring or message.
*/
type Decoder struct {
	rdr     io.Reader
	buf     [1024]byte // Staging area for yet-to-be-parsed bytes from io.Reader
	at, end int        // Current and last byte of buf not yet parsed

	parseError      error // Once a parse error has occurred, all bets are off forever
	state           parseState
	length          int    // Currently computed netstring length
	lengthValueRead int    // How many bytes of value have we read thus far?
	inProgress      []byte // The currently-being-parsed netstring
}

// NewDecoder constructs a Decoder which accepts a byte stream via its io.Reader interface
// and presents decoded netstrings via Decode(), DecodeKeyed() and Unmarshal()
func NewDecoder(rdr io.Reader) *Decoder {
	return &Decoder{rdr: rdr}
}

// parse picks up parsing from where it last left off and consumes bytes from the
// io.Reader until a complete netstring has been parsed. If an error is detected, parsing
// stops. Forever.
//
// The only tricky bit is parsing the length. There must be at least one length byte and a
// length can only start with a zero if that is the only length byte. That is, "0:," is
// valid, but "00:," is not nor is "01:A,". This is why there is a parseFirstByte state -
// to deal with leading zero constraints. Most netstring implementations do not bother
// checking this constraint, so it's possible that there may be interoperability issues
// with less fastidious netstring implementations.
//
// Any error is set in parse.Error but this should *only* be looked at if the returned
// netstring is nil. The reason for this slightly non idiomatic approach is that we want to
// make the error "sticky" *after* the error as it could be, e.g., io.EOF which should only
// be noticed after all bytes have been parsed.
func (dec *Decoder) parse() (good []byte) {
	if dec.parseError != nil {
		return
	}
	for { // Parse until error, EOF or netstring found
		if dec.at == dec.end { // Buffer empty?
			dec.end, dec.parseError = dec.rdr.Read(dec.buf[:])
			if dec.end == 0 { // dec.parseError better not be nil!
				return
			}
			dec.at = 0
		}

		var b byte
		for dec.at < dec.end {
			switch dec.state {

			case parseFirstByte: // Track leading zero
				b = dec.buf[dec.at]
				dec.at++
				if b < '0' || b > '9' { // A length digit?
					dec.parseError = ErrLengthNotDigit
					return
				}
				dec.length = int(b - '0')
				dec.state = parseLength

			case parseLength: // Second and subsequent length bytes
				b = dec.buf[dec.at]
				dec.at++
				if b >= '0' && b <= '9' { // A length digit?
					if dec.length == 0 {
						dec.parseError = ErrLeadingZero
						return
					}

					dec.length = dec.length*10 + int(b-'0')
					if dec.length > MaximumLength {
						dec.parseError = ErrLengthToLong
						return
					}
					continue
				}

				dec.state = parseColon
				fallthrough // "b" is still set and as yet unconsumed

			case parseColon:
				if b != leadingColon {
					dec.parseError = ErrColonExpected
					return
				}
				dec.inProgress = make([]byte, dec.length) // Container to return to caller
				dec.state = parseValue

			case parseValue:
				vr := dec.lengthValueRead // Current value length
				want := dec.length - vr   // How many bytes to complete the value?
				got := copy(dec.inProgress[vr:vr+want], dec.buf[dec.at:dec.end])
				dec.at += got
				dec.lengthValueRead += got
				if got == want { // Did we get all remaining bytes for this value?
					dec.state = parseComma // Yep, transition to next state
				}

			case parseComma:
				b = dec.buf[dec.at]
				dec.at++
				if b != trailingComma {
					dec.parseError = ErrCommaExpected
					return
				}

				// Have a good netstring, reset state and return netstring.

				good = dec.inProgress
				dec.inProgress = nil
				dec.state = parseFirstByte
				dec.length = 0
				dec.lengthValueRead = 0
				return
			}
		}
	}
}

// Decode returns the next available netstring. If no more netstrings are available from
// the supplied io.Reader, io.EOF is returned.
//
// Once an invalid netstring is detected, the byte stream is considered permanently
// unrecoverable and the same error is returned in perpetuity.
//
// The DecodeKeyed() function is better suited if the application is using "keyed"
// netstrings.
func (dec *Decoder) Decode() (ns []byte, err error) {
	ns = dec.parse()
	if ns != nil {
		return // Do not look at parseError until all netstrings consumed
	}

	err = dec.parseError

	return
}

// DecodeKeyed is used when the stream contains "keyed" netstrings created by the
// Encoder. A "keyed" netstring is simply a netstring where the first byte is a "key" used
// to categorize the rest of the value. What that categorization means is entirely up to
// the application.
//
// DecodeKeyed returns the next available netstring, if any, along with the prefix
// "key". The returned value does *not* include the prefix "key". If no more netstrings
// are available, error is returned with io.EOF.
//
// Once an invalid netstring is detected, the byte stream is considered permanently
// unrecoverable and the same error is returned in perpetuity.
//
// This function returns non-persistent errors if a non-keyed netstring is parsed. A
// non-keyed netstring is either zero length or the first byte is not an isalpha() key
// value.
func (dec *Decoder) DecodeKeyed() (Key, []byte, error) {
	ns := dec.parse()
	if ns == nil {
		return NoKey, nil, dec.parseError
	}

	if len(ns) == 0 { // No key byte is a temporary error
		return NoKey, nil, ErrZeroKey
	}

	key := Key(ns[0])
	keyed, err := key.Assess()
	if err != nil {
		return NoKey, nil, err
	}
	if !keyed { // Caller is expecting a "keyed" netstring
		return NoKey, nil, ErrInvalidKey
	}

	return key, ns[1:], nil
}
