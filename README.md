<!-- Always newline after period so diffs are easier to read. -->
# netstring - Robust encoding and decoding of netstrings

## Introduction

The `netstring` package is a go implementation of
[netstring](https://cr.yp.to/proto/netstrings.txt) serialization as originally specified
by D. J. Bernstein [djb](https://cr.yp.to/cv.html).

`netstring` provides an Encoder and a Decoder which are particularly suited to exchanging
netstrings across reliable transports such as TCP, WebSockets, D-Bus and the like.

A netstring.Encoder writes go types as encoded netstrings to a supplied io.Writer. Encoder
has helper functions to encode basic go types such as bool, ints, floats, strings and
bytes to netstrings. Structs, maps and other complex data structures are not supported.
A netstring.Decoder accepts a byte-stream via its io.Reader and presents successfully
parsed netstrings via the Decoder.Decode() and Decoder.DecodeKeyed() functions.

Alternatively applications can use the message level Encoder.Marshal() and
Decoder.Unmarshal() convenience functions to encode and decode a basic struct containing
"keyed" netstrings with an end-of-message sentinel.

The overall goal of this package is to make it easy to attach netstring Encoders and
Decoders to network connections or other reliable transports so that applications can
exchange messages with either the lower level Encode*() and Decode*() functions or the
higher level Marshal() and Unmarshal() functions.

### Project Status

[![Build Status](https://github.com/markdingo/netstring/actions/workflows/go.yml/badge.svg)](https://github.com/markdingo/netstring/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/markdingo/netstring/branch/main/graph/badge.svg)](https://codecov.io/gh/markdingo/netstring)
[![CodeQL](https://github.com/markdingo/netstring/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/markdingo/netstring/actions/workflows/codeql-analysis.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/markdingo/netstring)](https://goreportcard.com/report/github.com/markdingo/netstring)
[![Go Reference](https://pkg.go.dev/badge/github.com/markdingo/netstring.svg)](https://pkg.go.dev/github.com/markdingo/netstring)

## Background

A netstring is a serialization technique where each value is expressed in the form
[length] ":" [value] "," where [value] is the payload, [length] is the length of [value]
in decimal bytes and the colon and comma are leading and trailing delimiters respectively.
An example of the value "The Hitchhiker's Guide to the Galaxy - D.A." encoded as a
netstring is:

    "42:The Hitchhiker's Guide to the Galaxy - DA.,"

Storing binary values in netstrings, while possible, is not recommended for obvious
reasons of incompatible CPU architectures.
Best practise is to convert all binary values to strings prior to encoding.
To assist in this best practice, helper functions are available to encode basic go-types
such as ints and floats to netstrings.
E.g. the function Encoder.EncodeInt() converts int(2^16) to the netstring:

    "5:65536,"

## "Keyed" netstrings

In addition to standard netstrings, this package supports "keyed" netstrings which are
nothing more than netstrings where the first byte of [value] signifies a "type" which
describes the rest of [value] in some useful way to the application.

The benefit of "keyed" netstrings is that they create a self-describing typing system
which allows applications to encode multiple netstrings in a message without having to
rely on positional order or nested netstrings to convey semantics or encapsulate a message.

To demonstrate the benefits of "keyed" netstrings, say you want to encode Name, Age,
Country and Height as netstrings to be transmitted to a remote service? With standard
netstrings you'd have to agree on the positional order for each value, say netstring #0
for Name, #1 for Age and so on.

Here is what that series of netstrings looks like:

    "4:Name,3:Age,7:Country,6:Height,"

If any of these values are optional, you'd still have to provide a zero length netstring
to maintain positional order; however this creates the ambiguity as to whether the intent
is to convey a zero length string or a NULL.

With "keyed" netstrings the values can be presented in any order and optional (or NULL)
values are simply omitted.
In the above example if we assign the "key" of 'n' to Name, 'a' to Age, 'c' to Country and
'h' to Height the series of "keyed" netstrings looks like:

    "5:nName,4:aAge,8:cCountry,7:hHeight,"

and if Age is optional the series of netstrings simply becomes:

    "8:cCountry,7:hHeight,5:nName,"

Note the change of order as well as the missing 'a' netstring?
All perfectly acceptable with "keyed" netstrings.

To gain the most benefit from "keyed" netstrings, the usual strategy is to reserve a
specific key value as an "end-of-message" sentinal which naturally is the last netstring
in the message.
The convention is to use 'z' as the "end-of-message" sentinal as demonstrated in
subsequent examples.

## Installation

When imported by your program, `github.com/markdingo/netstring` should automatically
install with `go mod tidy`.

Once installed, you can run the package tests with:

```
 go test github.com/markdingo/netstring
```

as well as display the package documentation with:

```
 go doc github.com/markdingo/netstring
```

## Usage and Examples

``` go
import "github.com/markdingo/netstring"
```

To create a message of netstrings, call NewEncoder() then call the various Encode*()
functions to encode the basic go types.
This code fragment encodes a message into a bytes.Buffer.

```
 var buf bytes.Buffer
 enc := netstring.NewEncoder(&buf)
 enc.EncodeInt('a', 21)           // Age
 enc.EncodeString('C', "Iceland") // Country
 enc.EncodeString('n', "Bjorn")   // Name
 enc.EncodeBytes('z')             // End-of-message sentinel
 fmt.Println(buf.String())        // "3:a21,8:CIceland,6:nBjorn,1:z,"
```

And this fragment decodes the same message.

```
 dec := netstring.NewDecoder(&buf)
 k, v, e := dec.DecodeKeyed() // k=a, v=21
 k, v, e = dec.DecodeKeyed()  // k=C, v=Iceland
 k, v, e = dec.DecodeKeyed()  // k=n, v=Bjorn
 k, v, e = dec.DecodeKeyed()  // k=z End-Of-Message
```

The message is more conveniently encoded with Marshal() as this fragment shows:

```
 type record struct {
      Age         int    `netstring:"a"`
      Country     string `netstring:"c"`
      Name        string `netstring:"n"`
 }

var buf bytes.Buffer
 enc := netstring.NewEncoder(&buf)
 out := &record{21, "Iceland", "Bjorn"}
 enc.Marshal('z', out)
```

and more conveniently decoded with Unmarshal() as this fragment shows:

```
 dec := netstring.NewDecoder(&buf)
 in := &record{}
 dec.Unmarshal('z', in)
```

Full working programs can be found in the _examples directory.

### Community

If you have any problems using `netstring` or suggestions on how it can do a better job,
don't hesitate to create an [issue](https://github.com/markdingo/netstring/issues) on the
project home page.
This package can only improve with your feedback.

### Copyright and License

`netstring` is Copyright :copyright: 2023 Mark Delany and is licensed under the BSD
2-Clause "Simplified" License.
