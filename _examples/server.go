package main

import (
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/markdingo/netstring"
)

// Demonstrate encoders and decoders automatically encoding and decoding over a network
// connection.
//
// The pipeline looks like this:
//
// Client -> Encoder -> net.Conn -> Decoder -> Server ..v
//                                                      v
//                                                      v
//    ... <- Decoder <- net.Conn <- Encoder          <..v
//
// Client C sends request messages to the Server who response with a response messages.
//
// Request message is a Plus Mode netstring where:
//
// f = function
// i = input
// z = End-Of-Message
//
// Valid functions are ToLower and ToUpper.
//
// The response message is:
//
// e = optional error
// o = output
// z = End-Of-Message

const (
	functionKey = 'f'
	inputKey    = 'i'
	outputKey   = 'o'
	errorKey    = 'e'
	EOMKey      = 'z'

	lowerFunction = "ToLower"
	upperFunction = "ToUpper"
)

func main() {
	ln, err := net.Listen("tcp", ":8086")
	if err != nil {
		fmt.Println(err)
		return
	}

	conn, err := ln.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}
	dec := netstring.NewDecoder(conn) // For requests from client
	enc := netstring.NewEncoder(conn) // For sending responses to client

	fmt.Println("Server Ready")
	var function, input string
	for {
		k, v, e := dec.DecodeKeyed()
		if e != nil {
			if e == io.EOF {
				fmt.Println("Server got EOF")
			} else {
				fmt.Println("Fatal Server Error", e)
			}
			return
		}

		switch k {
		case functionKey:
			function = string(v)
		case inputKey:
			input = string(v)
		case EOMKey:
			fmt.Println("Server Request", function, input)
			switch function {
			case lowerFunction:
				enc.EncodeString(outputKey, strings.ToLower(input))
			case upperFunction:
				enc.EncodeString(outputKey, strings.ToUpper(input))
			default:
				enc.EncodeString(errorKey, "Invalid function "+function)
			}
			enc.EncodeBytes(EOMKey)
			function = ""
			input = ""
		default:
			fmt.Println("Invalid Key", byte(k))
			return
		}
	}
	fmt.Println("Server Done")
}
