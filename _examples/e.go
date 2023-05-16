package main

import (
	"bytes"
	"fmt"

	"github.com/markdingo/netstring"
)

// This program contains all the example code in doc.go etc. It is not checked in to the
// repo but it should not be deleted.
func main() {
	fmt.Println("doc1")
	doc1()
	fmt.Println("doc2")
	doc2()
	fmt.Println("marshal")
	marshal()
	fmt.Println("unmarshal")
	unmarshal()
	fmt.Println("readme1")
	readme1()
	fmt.Println("readme2")
	readme2()
}

func doc1() {
	var buf bytes.Buffer
	enc := netstring.NewEncoder(&buf)
	enc.EncodeInt('a', 21)           // Age
	enc.EncodeString('C', "Iceland") // Country
	enc.EncodeString('n', "Bjorn")   // Name
	enc.EncodeBytes('z')             // End-of-message sentinel
	fmt.Println(buf.String())        // "3:a21,8:CIceland,6:nBjorn,1:z,"

	dec := netstring.NewDecoder(&buf)
	k, v, e := dec.DecodeKeyed() // k=a, v=21
	k, v, e = dec.DecodeKeyed()  // k=C, v=Iceland
	k, v, e = dec.DecodeKeyed()  // k=n, v=Bjorn
	k, v, e = dec.DecodeKeyed()  // k=z End-Of-Message
	_ = k
	_ = v
	_ = e
}

func doc2() {
	type msg struct {
		Age     int    `netstring:"a"`
		Country string `netstring:"C"`
		Name    string `netstring:"n"`
	}

	var buf bytes.Buffer
	enc := netstring.NewEncoder(&buf)
	out := &msg{21, "Iceland", "Bjorn"}
	enc.Marshal('z', out)
	fmt.Println(buf.String()) // "3:a21,8:CIceland,6:nBjorn,1:z,"

	dec := netstring.NewDecoder(&buf)
	in := &msg{}
	dec.Unmarshal('z', in)
	fmt.Println(in)
}

func marshal() {
	type record struct {
		Age         int    `netstring:"a"`
		Country     string `netstring:"c"`
		TLD         []byte `netstring:"t"`
		CountryCode []byte `netstring:"C"`
		Name        string `netstring:"n"`
		Height      uint16 // Ignored - no netstring tag
		dbKey       int64  // Ignored - not exported
	}

	var bbuf bytes.Buffer
	enc := netstring.NewEncoder(&bbuf)
	enc.EncodeString('M', "r0") // Message type 'r', version zero

	r := record{22, "New Zealand", []byte{'n', 'z'}, []byte("64"), "Bob", 173, 42}
	err := enc.Marshal('Z', &r)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(bbuf.String()) // "3:Mr0,3:a22,12:cNew Zealand,3:tnz,3:C64,4:nBob,1:Z,"
}

func unmarshal() {
	type record struct {
		Age         int    `netstring:"a"`
		Country     string `netstring:"c"`
		TLD         []byte `netstring:"t"`
		CountryCode []byte `netstring:"C"`
		Name        string `netstring:"n"`
	}

	bbuf := bytes.NewBufferString("3:Mr0,3:a22,11:cNew Zeland,3:C64,4:nBob,1:Z,")
	dec := netstring.NewDecoder(bbuf)
	k, v, _ := dec.DecodeKeyed()
	if k == 'M' && string(v) == "r0" { // Dispatch on message type
		msg := &record{}
		dec.Unmarshal('Z', msg)
		fmt.Println(msg)
	}
}

func readme1() {
	var buf bytes.Buffer
	enc := netstring.NewEncoder(&buf)
	enc.EncodeInt('a', 21)           // Age
	enc.EncodeString('C', "Iceland") // Country
	enc.EncodeString('n', "Bjorn")   // Name
	enc.EncodeBytes('z')             // End-of-message sentinel
	fmt.Println(buf.String())        // "3:a21,8:CIceland,6:nBjorn,1:z,"

	dec := netstring.NewDecoder(&buf)
	k, v, e := dec.DecodeKeyed() // k=a, v=21
	k, v, e = dec.DecodeKeyed()  // k=C, v=Iceland
	k, v, e = dec.DecodeKeyed()  // k=n, v=Bjorn
	k, v, e = dec.DecodeKeyed()  // k=z End-Of-Message
	_ = k
	_ = v
	_ = e
}

func readme2() {
	type record struct {
		Age     int    `netstring:"a"`
		Country string `netstring:"c"`
		Name    string `netstring:"n"`
	}

	var buf bytes.Buffer
	enc := netstring.NewEncoder(&buf)
	out := &record{21, "Iceland", "Bjorn"}
	enc.Marshal('z', out)
	fmt.Println(buf.String())

	dec := netstring.NewDecoder(&buf)
	in := &record{}
	dec.Unmarshal('z', in)
	fmt.Println(in)
}
