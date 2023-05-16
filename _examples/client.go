package main

import (
	"fmt"
	"net"

	"github.com/markdingo/netstring"
)

// See server.go for documentation.

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
	conn, err := net.Dial("tcp", ":8086")
	if err != nil {
		fmt.Println(err)
		return
	}

	enc := netstring.NewEncoder(conn)
	dec := netstring.NewDecoder(conn)

	fmt.Println("Client Ready")
	basicEncoderDecoder(enc, dec)
	mashalEncoderDecoder(enc, dec)
	conn.Close()
	fmt.Println("Client Done")
}

// Use the basic Encode*() functions to send each individual netstring
func basicEncoderDecoder(enc *netstring.Encoder, dec *netstring.Decoder) {
	for _, input := range []string{"aaaaa", "bb", "ccccc", "dddddd"} {
		e := enc.EncodeString(functionKey, upperFunction)
		if e != nil {
			panic(e)
		}
		e = enc.EncodeString(inputKey, input)
		if e != nil {
			panic(e)
		}
		e = enc.EncodeBytes(EOMKey)
		if e != nil {
			panic(e)
		}
		fmt.Println("Client sending", input)
		var error, output string
		done := false
		for !done {
			k, v, e := dec.DecodeKeyed()
			if e != nil {
				fmt.Println("Client GetKeyed Error", e)
				return
			}
			switch k {
			case errorKey:
				error = string(v)
			case outputKey:
				output = string(v)
			case EOMKey:
				done = true
			default:
				fmt.Println("Invalid netstring key", byte(k))
				return
			}
		}
		fmt.Println("Client got", output, error)
	}
}

// Use the higher level marshal functions to exchange messages
func mashalEncoderDecoder(enc *netstring.Encoder, dec *netstring.Decoder) {
	type request struct {
		Function string `netstring:"f"`
		Input    string `netstring:"i"`
	}
	type response struct {
		Output string `netstring:"o"`
		Error  string `netstring:"e"`
	}

	for _, input := range []string{"Turtle", "waxy", "belief", "SomeAreUpper"} {
		fmt.Println("Client sending", input)
		req := &request{upperFunction, input}
		err := enc.Marshal(EOMKey, req)
		if err != nil {
			fmt.Println(err)
			return
		}

		resp := &response{}
		unknown, err := dec.Unmarshal(EOMKey, resp)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Client got", unknown.String(), resp.Output, resp.Error)
	}
}
