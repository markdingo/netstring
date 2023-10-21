package netstring_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/markdingo/netstring"
)

type benchNullWriter struct {
}

func (w *benchNullWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

// Baseline performance
func BenchmarkEncodeBaseline(b *testing.B) {
	sa := []byte{'a', 'b', 'c', 'd', 'e', 'a', 'b', 'c', 'd', 'e'}

	enc := netstring.NewEncoder(&benchNullWriter{})
	for i := 0; i < b.N; i++ {
		err := enc.EncodeBytes('A', sa)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Shows cost of range iteration
func BenchmarkEncodeBytes(b *testing.B) {
	sa := []byte{'a'}
	sb := []byte{'b'}
	sc := []byte{'c'}
	sd := []byte{'d'}
	se := []byte{'e'}

	enc := netstring.NewEncoder(&benchNullWriter{})
	for i := 0; i < b.N; i++ {
		err := enc.EncodeBytes('A', sa, sb, sc, sd, se, sa, sb, sc, sd, se)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Compare with Baseline to show cost of wrapped function and string conversion
func BenchmarkEncodeString1(b *testing.B) {
	enc := netstring.NewEncoder(&benchNullWriter{})
	s := "abcdeabcde"
	for i := 0; i < b.N; i++ {
		err := enc.EncodeString('A', s)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Get a relative cost of Encoding vs Decoding
func BenchmarkCompareEncode(b *testing.B) {
	wBuf := bytes.NewBuffer(make([]byte, 0, 200))
	enc := netstring.NewEncoder(wBuf)
	ab := []byte{'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a'}
	bb := []byte{'b', 'b', 'b', 'b', 'b', 'b', 'b', 'b', 'b', 'b'}
	cb := []byte{'c', 'c', 'c', 'c', 'c', 'c', 'c', 'c', 'c', 'c'}
	db := []byte{'d', 'd', 'd', 'd', 'd', 'd', 'd', 'd', 'd', 'd'}
	eb := []byte{'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e'}
	fb := []byte{'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f'}
	gb := []byte{'g', 'g', 'g', 'g', 'g', 'g', 'g', 'g', 'g', 'g'}
	hb := []byte{'h', 'h', 'h', 'h', 'h', 'h', 'h', 'h', 'h', 'h'}
	ib := []byte{'i', 'i', 'i', 'i', 'i', 'i', 'i', 'i', 'i', 'i'}
	jb := []byte{'j', 'j', 'j', 'j', 'j', 'j', 'j', 'j', 'j', 'j'}
	for i := 0; i < b.N; i++ {
		wBuf.Reset()
		enc.EncodeBytes('A', ab)
		enc.EncodeBytes('B', bb)
		enc.EncodeBytes('C', cb)
		enc.EncodeBytes('D', db)
		enc.EncodeBytes('E', eb)
		enc.EncodeBytes('F', fb)
		enc.EncodeBytes('G', gb)
		enc.EncodeBytes('H', hb)
		enc.EncodeBytes('I', ib)
		enc.EncodeBytes('J', jb)
		enc.EncodeBytes('z')
	}
}

// Decoding is less expensive than encoding - which is a bit of a surprise.
func BenchmarkCompareDecode(b *testing.B) {
	var wBuf bytes.Buffer
	enc := netstring.NewEncoder(&wBuf)
	enc.EncodeString('A', "aaaaaaaaaa")
	enc.EncodeString('B', "bbbbbbbbbb")
	enc.EncodeString('C', "cccccccccc")
	enc.EncodeString('D', "dddddddddd")
	enc.EncodeString('E', "eeeeeeeeee")
	enc.EncodeString('F', "ffffffffff")
	enc.EncodeString('G', "gggggggggg")
	enc.EncodeString('H', "hhhhhhhhhh")
	enc.EncodeString('I', "iiiiiiiiii")
	enc.EncodeString('J', "jjjjjjjjjj")
	enc.EncodeBytes('z')
	rBuf := bytes.NewReader(wBuf.Bytes())
	for i := 0; i < b.N; i++ {
		rBuf.Seek(0, io.SeekStart)
		dec := netstring.NewDecoder(rBuf)
		for j := 'A'; j < 'J'; j++ {
			k, buf, err := dec.DecodeKeyed()
			if err != nil {
				b.Fatal(err)
			}
			if int(k) != int(j) {
				b.Fatal("Wrong Key", j, k)
			}
			if len(buf) != 10 {
				b.Fatal("Wrong length. Expected 10, got", len(buf))
			}
		}
	}
}
