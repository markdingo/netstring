package netstring

// Key is the byte value provided to the Encoder Encode*() functions to determine whether
// the encoded netstring is a standard netstring or a "keyed" netstring. Valid values are:
// NoKey (or 0) for a standard netstring or an isalpha() value ('a'-'z' or 'A'-'Z') for a
// "keyed" netstring. All other values are invalid. Key is also the type returned by the
// Decoder functions. Use Key.Assess() to determine the validity and type of a key.
type Key byte

// NoKey is the special "key" provided to the Encoder.Encode*() functions to indicate that
// a standard netstring should be encoded.
const NoKey Key = 0

func (k Key) String() string {
	return string(k)
}

// Assess determines whether the Key 'k' is valid or not and whether it implies a standard
// or "keyed" netstring. NoKey.Assess() returns keyed=false and err=nil which is to say
// that Assess treats NoKey as valid but it signifies a standard netstring.
//
// "keyed" is set true if 'k' is in the range 'a'-'z' or 'A'-'Z', inclusive.
func (k Key) Assess() (keyed bool, err error) {
	if k == NoKey {
		return false, nil
	}
	if (k >= 'a' && k <= 'z') || (k >= 'A' && k <= 'Z') {
		return true, nil
	}

	return false, ErrInvalidKey
}
