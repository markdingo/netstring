package netstring

import (
	"fmt"
	"reflect"
)

// Marshal takes "message" as a simple struct or a pointer to a simple struct and encodes
// all exported fields with a "netstring" tag as a series of "keyed" netstrings. If there
// is no "netstring" tag the field is ignored. The reason the "netstring" tag is required
// is to supply a netstring key value which assists Unmarshal in locating the appropriate
// field on the receiving side. Marshal cannot be used to encode standard netstrings.
//
// The "eom" parameter is used to create an end-of-message sentinel "keyed" netstring. It
// can be any valid Key excepting netstring.NoKey. The sentinel follows the simple struct
// netstrings with Encoder.EncodeBytes(eom).
//
// There are significant constraints as to what constitutes a valid simple struct. In
// large part this is because netstrings are ill-suited to support complex messages - use
// encoding/json or protobufs for those. Candidate fields (i.e. exported with a
// "netstring" tag) can only be one of the following basic go types: all ints and uints,
// all floats, strings and byte slices. That's it! Put another way, fields cannot be
// complex types such as maps, arrays, structs, pointers, etc. Any unsupported field type
// which has a "netstring" tag returns an error.
//
// The "netstring" tag value must be a valid netstring.Key and each "netstring" tag value
// must be unique otherwise an error is returned.
//
// Though fields are encoded in the order found in the struct via the "reflect" package,
// this sequence should not be relied on. Always use the "keyed" values to associate
// netstrings to fields.
//
// To assist go applications wishing to Unmarshal, it is good practice to use the first
// netstring as a message type so that the receiving side can select the corresponding
// struct to Unmarshal in to. Having to know the type before seeing the payload is a
// fundamental issue for all go Unmarshal functions such as json.Unmarshal in that they
// have to know ahead of time what type of struct the message contains; thus the message
// type has to effectively preceed the message. At least with netstrings that's easy to
// arrange.
//
// Type and tag checking is performed while encoding so any error return probably leaves
// the output stream in an indeterminate state.
//
// An example:
//
//	type record struct {
//	  Age         int     `netstring:"a"`
//	  Country     string  `netstring:"c"`
//	  TLD         []byte  `netstring:"t"`
//	  CountryCode []byte  `netstring:"C"`
//	  Name        string  `netstring:"n"`
//	  Height      uint16  // Ignored - no netstring tag
//	  dbKey       int64   // Ignored - not exported
//	}
//	...
//	var bbuf bytes.Buffer
//	enc := netstring.NewEncoder(&bbuf)
//	enc.EncodeString('M', "r0")   // Message type 'r', version zero
//
//	r := record{22, "New Zealand", []byte{'n', 'z'}, []byte("64"), "Bob", 173, 42}
//	enc.Marshal('Z', &r)
//
//	fmt.Println(bbuf.String()) // "3:Mr0,3:a22,12:cNew Zealand,3:tnz,3:C64,4:nBob,1:Z,"
//
// Particularly note the preceeding message type "r0" and the trailing end-of-message
// sentinel 'Z'.
func (enc *Encoder) Marshal(eom Key, message any) error {
	k, e := eom.Assess()
	if e != nil {
		return e
	}
	if !k {
		return ErrBadMarshalEOM
	}

	vo := reflect.ValueOf(message) // vo is a reflect.Value
	if !vo.IsValid() {
		return ErrBadMarshalValue
	}
	to := vo.Type()
	kind := vo.Kind()
	if kind == reflect.Pointer { // If it's a pointer, step into the struct
		vo = vo.Elem()
		to = vo.Type()
		kind = vo.Kind()
	}
	if kind != reflect.Struct { // Only go one-level deep, so no **struct{}
		return ErrBadMarshalValue
	}

	dupes := make(map[Key]string)
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
			return fmt.Errorf("%s%s tag '%s' (0x%X) is not a valid netstring.Key",
				errorPrefix, sf.Name, tag, tag)
		}
		key := Key(tag[0])
		keyed, err := key.Assess()
		if err != nil {
			return err
		}
		if !keyed {
			return fmt.Errorf("%s%s tag '%s' (0x%X) is not a valid netstring.Key",
				errorPrefix, sf.Name, tag, tag)
		}
		if n, ok := dupes[key]; ok {
			return fmt.Errorf("%sDuplicate tag '%s' for '%s' and '%s'",
				errorPrefix, tag, sf.Name, n)
		}
		dupes[key] = sf.Name

		kind := sf.Type.Kind()
		vf := vo.Field(ix)
		switch kind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			enc.EncodeInt64(key, vf.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			enc.EncodeUint64(key, vf.Uint())
		case reflect.Float32, reflect.Float64:
			enc.EncodeFloat64(key, vf.Float())
		case reflect.String:
			enc.EncodeString(key, vf.String())
		case reflect.Slice: // Is it a byte slice?
			eKind := sf.Type.Elem().Kind()
			if eKind == reflect.Uint8 {
				enc.EncodeBytes(key, vf.Bytes())
			} else {
				return fmt.Errorf("%s%s type unsupported (%s of %s)",
					errorPrefix, sf.Name, kind, eKind)
			}

		default:
			return fmt.Errorf("%s%s type unsupported (%s)",
				errorPrefix, sf.Name, kind)
		}
	}

	enc.EncodeBytes(eom)

	return nil
}
