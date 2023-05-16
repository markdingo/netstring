package main

// This is the complete program as highlighted in the package docs. It encodes a simple
// message with "keyed" netstrings then decodes the same message then compares the before
// and after structs to ensure equality.

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/markdingo/netstring"
)

type record struct {
	age     int
	name    string
	country string
}

func main() {
	r1 := record{21, "Bjorn", "Iceland"}
	var buf bytes.Buffer
	ec := netstring.NewEncoder(&buf)

	ec.EncodeInt('a', r1.age)        // Using type-specific encoders for int
	ec.EncodeString('C', r1.country) // and string
	ec.Encode('n', r1.name)          // The generic encoder
	ec.EncodeBytes('z')              // End of message sentinal
	message := buf.String()
	fmt.Println(message) // "3:a21,8:CIceland,6:nBjorn,1:z,"

	// Decode the message to reconstruct the same record

	var r2 record
	dc := netstring.NewDecoder(&buf)

	eom := false
	for !eom {
		k, v, _ := dc.DecodeKeyed()
		s := string(v)
		switch k {
		case 'a':
			r2.age, _ = strconv.Atoi(s)
		case 'C':
			r2.country = s
		case 'n':
			r2.name = s
		case 'z':
			eom = true
		}
	}

	if r1 != r2 {
		fmt.Println("Compare: bad")
	} else {
		fmt.Println("Compare: good")
	}
}
