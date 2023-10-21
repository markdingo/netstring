package netstring_test

import (
	"bytes"
	"io"
	"strconv"
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

type bmStruct struct {
	Age         int    `netstring:"a"`
	Country     string `netstring:"c"`
	TLD         []byte `netstring:"t"`
	CountryCode []byte `netstring:"C"`
	Name        string `netstring:"n"`
	Height      uint16 `netstring:"H"`
	Key         int64  `netstring:"K"`
}

// The higher-level Marshal function turns out to be about 4 times slower than manually encoding.
func BenchmarkMarshalAuto(b *testing.B) {
	s := bmStruct{21, "Iceland", []byte{'i', 'c'}, []byte("354"), "Bjorn", 183, 123456}
	enc := netstring.NewEncoder(&benchNullWriter{})
	for i := 0; i < b.N; i++ {
		err := enc.Marshal('Z', s)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Do what MarshalAuto does, but using the simpler and lower-level Encoders directly. The
// extra cost of Marshal is largely due to the tag fetching, deduping and reflection
// aspects.
func BenchmarkMarshalManual(b *testing.B) {
	s := bmStruct{21, "Australia", []byte{'a', 'u'}, []byte("354"), "Bruce", 200, 987654}
	enc := netstring.NewEncoder(&benchNullWriter{})
	for i := 0; i < b.N; i++ {
		err := enc.EncodeInt('a', s.Age)
		if err != nil {
			b.Fatal(err)
		}
		err = enc.EncodeString('c', s.Country)
		if err != nil {
			b.Fatal(err)
		}
		err = enc.EncodeBytes('t', s.TLD)
		if err != nil {
			b.Fatal(err)
		}
		err = enc.EncodeBytes('C', s.CountryCode)
		if err != nil {
			b.Fatal(err)
		}
		err = enc.EncodeString('n', s.Name)
		if err != nil {
			b.Fatal(err)
		}
		err = enc.EncodeUint('H', uint(s.Height))
		if err != nil {
			b.Fatal(err)
		}
		err = enc.EncodeInt64('K', s.Key)
		if err != nil {
			b.Fatal(err)
		}
		err = enc.EncodeBytes('Z')
		if err != nil {
			b.Fatal(err)
		}
	}
}

// As with marshal, the higher levvel Unmarshal turns out to be about 3-4 times slower
// than manually decoding.
func BenchmarkUnmarshalAuto(b *testing.B) {
	in := "3:a99,10:cAustralia,3:tau,4:C354,6:nBruce,4:H200,7:K987654,1:Z,"
	rBuf := bytes.NewReader([]byte(in))
	for i := 0; i < b.N; i++ {
		rBuf.Seek(0, io.SeekStart)
		dec := netstring.NewDecoder(rBuf)
		var s bmStruct
		unk, err := dec.Unmarshal('Z', &s)
		if err != nil {
			b.Fatal("iter", i, "unmarshal returned", err)
		}
		if unk != 0 {
			b.Fatal("Unknown key returned", unk)
		}
	}
}

// Do what UnmarshalAuto does, but using the simpler and lower-level Decoders directly.
func BenchmarkUnmarshalManual(b *testing.B) {
	in := "3:a99,10:cAustralia,3:tau,4:C354,6:nBruce,4:H200,7:K987654,1:Z,"
	rBuf := bytes.NewReader([]byte(in))
	for i := 0; i < b.N; i++ {
		rBuf.Seek(0, io.SeekStart)
		dec := netstring.NewDecoder(rBuf)
		var s bmStruct
		for {
			k, v, e := dec.DecodeKeyed()
			if e != nil {
				b.Fatal(e)
			}

			if k == 'Z' {
				break
			}

			switch k {
			case 'a':
				s.Age, _ = strconv.Atoi(string(v))
			case 'c':
				s.Country = string(v)
			case 't':
				s.TLD = v
			case 'C':
				s.CountryCode = v
			case 'n':
				s.Name = string(v)
			case 'H':
				u16, _ := strconv.Atoi(string(v))
				s.Height = uint16(u16)
			case 'K':
				i64, _ := strconv.Atoi(string(v))
				s.Key = int64(i64)
			default:
				b.Fatal("Unexpected Key type", k)
			}
		}
	}
}
