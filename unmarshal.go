package netstring

import (
	"fmt"
	"reflect"
	"strconv"
)

// Unmarshal takes incoming "keyed" netstrings and populates "message". Message must be a
// pointer to a simple struct with the same restrictions as discussed in Marshal.
//
// Each netstring is read via Decoder.DecodeKeyed() until a "keyed" netstring matches
// "eom". Each netstring is decoded into the field with a "netstring" tag matching the
// netstring "key".
//
// The end-of-message sentinel, "eom", can be any valid Key excepting
// netstring.NoKey. When the "eom" netstring is seen, the message is considered fully
// populated, the "eom" message is discarded and control is returned to the caller.
//
// If "message" is not a simple struct or pointer to a simple struct an error is returned.
// Only exported fields with "netstring" tags are considered for incoming "keyed"
// netstrings. If "message" contains duplicate "netstring" tag values an error is
// returned.
//
// The "unknown" variable is set with the key of any incoming "keyed" netstring which has
// no corresponding field in "message". Obviously only one "unknown" is visible to the
// caller even though there may be multiple occurrences. Since an unknown key may be
// acceptable to the application, it is left to the caller to decide whether this
// situation results in an error, an alert to upgrade, or silence.
//
// An example:
//
//	type record struct {
//	  Age         int     `netstring:"a"`
//	  Country     string  `netstring:"c"`
//	  TLD         []byte  `netstring:"t"`
//	  CountryCode []byte  `netstring:"C"`
//	  Name        string  `netstring:"n"`
//	}
//
//	bbuf := bytes.NewBufferString("3:Mr0,3:a22,11:cNew Zeland,3:C64,4:nBob,1:Z,")
//	dec := netstring.NewDecoder(bbuf)
//	k, v, e := dec.DecodeKeyed()
//	if k == 'M' && string(v) == "r0" {    // Dispatch on message type
//	   msg := &record{}
//	   dec.Unmarshal('Z', msg)
//	}
//
// Note how the first netstring is used to determine which struct to Unmarshal into.
func (dec *Decoder) Unmarshal(eom Key, message any) (unknown Key, err error) {
	k, e := eom.Assess()
	if e != nil {
		err = e
		return
	}
	if !k {
		err = ErrBadMarshalEOM
		return
	}

	vo := reflect.ValueOf(message) // vo is a reflect.Value
	if !vo.IsValid() {
		err = ErrBadMarshalValue
		return
	}
	to := vo.Type()
	kind := vo.Kind()
	if kind != reflect.Pointer { // Must be a Pointer so we can set it!
		err = ErrBadUnmarshalMsg
		return
	}

	vo = vo.Elem()
	to = vo.Type()
	kind = vo.Kind()

	if kind != reflect.Struct { // Only go one-level deep, so no **struct{}
		err = ErrBadUnmarshalMsg
		return
	}

	// Evaluate message fields

	type field struct {
		seen   bool
		name   string
		kind   reflect.Kind
		value  reflect.Value
		maxint int64
	}
	keyToField := make(map[Key]*field)

	for ix := 0; ix < to.NumField(); ix++ {
		sf := to.Field(ix) // Get StructField
		if !sf.IsExported() {
			continue
		}
		tag := sf.Tag.Get("netstring")
		if len(tag) == 0 {
			continue
		}
		if len(tag) != 1 {
			err = fmt.Errorf("%s%s tag '%s' (0x%X) is not a single character",
				errorPrefix, sf.Name, tag, tag)
			return
		}
		key := Key(tag[0])
		var keyed bool
		keyed, err = key.Assess()
		if err != nil {
			return
		}
		if !keyed {
			err = fmt.Errorf("%s%s tag '%s' (0x%X) is not a valid netstring.Key",
				errorPrefix, sf.Name, tag, tag)
			return
		}
		if f, ok := keyToField[key]; ok {
			err = fmt.Errorf("%sDuplicate tag '%s' for '%s' and '%s'",
				errorPrefix, tag, sf.Name, f.name)
			return
		}

		vf := vo.Field(ix)
		kind := sf.Type.Kind()

		// Some kinds need further checking
		switch kind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64: // Do nothing
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64: // Do nothing
		case reflect.Float32, reflect.Float64: // Do nothing
		case reflect.String: // Do nothing

		case reflect.Slice: // Is it a byte slice?
			eKind := sf.Type.Elem().Kind()
			if eKind != reflect.Uint8 {
				err = fmt.Errorf("%s%s type unsupported (%s of %s)",
					errorPrefix, sf.Name, kind, eKind)
				return
			}

		default:
			err = fmt.Errorf("%s%s type unsupported (%s)",
				errorPrefix, sf.Name, kind)
			return
		}

		keyToField[key] = &field{false, sf.Name, kind, vf, 0} // field looks good, stash it in the map
	}

	// Have all the information about message destination fields so start consuming
	// keyed netstrings and map them into the simple struct destination fields.

	for {
		k, v, e := dec.DecodeKeyed()
		if e != nil {
			err = e
			return
		}

		if k == eom {
			return
		}

		field, ok := keyToField[k]
		if !ok {
			unknown = k
			continue
		}

		if field.seen {
			err = fmt.Errorf("%sDuplicate key '%s' in decode stream for %s",
				errorPrefix, k.String(), field.name)
			return
		}
		field.seen = true

		switch field.kind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			vi, e := strconv.ParseInt(string(v), 10, 64)
			if e != nil || field.value.OverflowInt(vi) {
				err = fmt.Errorf("%sCannot convert '%s' to int for %s (%s)",
					errorPrefix, string(v), field.name, field.kind)
				return
			}
			field.value.SetInt(vi)

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			vi, e := strconv.ParseUint(string(v), 10, 64)
			if e != nil || field.value.OverflowUint(vi) {
				err = fmt.Errorf("%sCannot convert '%s' to uint for %s - overflows %s",
					errorPrefix, string(v), field.name, field.kind)
				return
			}
			field.value.SetUint(vi)

		case reflect.Float32, reflect.Float64:
			vf, e := strconv.ParseFloat(string(v), 64)
			if e != nil || field.value.OverflowFloat(vf) {
				err = fmt.Errorf("%sCannot convert '%s' to float for %s - overflows %s",
					errorPrefix, string(v), field.name, field.kind)
				return
			}
			field.value.SetFloat(vf)

		case reflect.String:
			field.value.SetString(string(v))

		case reflect.Slice:
			field.value.SetBytes(v)

		default:
			err = fmt.Errorf("%s%s Internal Error type (%s) ducked early check",
				errorPrefix, field.name, kind)
		}
	}
}
