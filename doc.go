/*
Package netstring provides robust encoding and decoding of netstrings to and from byte
streams.

netstrings are a simple serialization technique originally defined by
http://cr.yp.to/netstrings.txt. Typical usage is to exchange messages consisting of a
small number of well-defined netstrings. Complex messages with many variables and changing
semantics are better suited to more sophisticated encoding schemes such as encoding/json
or Protobufs.

Netstrings are of the form: [length] ":" [value] "," where [value] is the payload of
interest, [length] is the length of [value] in decimal bytes and the colon and comma are
leading and trailing delimiters respectively. The string "The Hitchhiker's Guide to the
Galaxy - D.A."  encoded as a netstring is:

	"42:The Hitchhiker's Guide to the Galaxy - DA.,"

# Binary Values

Storing binary values in netstrings, while possible, is not recommended for obvious
reasons of incompatible CPU architectures. Best practise is to convert all binary values
to strings prior to encoding.

To assist in this best practice, helper functions are available to encode basic go-types
such as ints and floats to netstrings. E.g. the function Encoder.EncodeInt() converts
int(2^16) to the netstring:

	"5:65536,"

Most of these helpers use strconv.Format* functions to convert binary values to strings
and applications are encouraged to use the corresponding strconv.Parse*() functions to
decode non-string values back to internal binary. The specifics of each to non-string
conversion are documented in each helper function.

Apart from simple struct support with Marshal() and Unmarshal() there is no support for
encoding complex go types such as nest structs, arrays, slices and maps as this is the
juncture at which the application might best be served using a more sophisticated encoding
scheme as mentioned earlier.

# Rigorous Parsing

This package is particularly fastidious about parsing and generating valid netstrings. For
example, the specification say that a length can only start with a zero digit if the
length field is exactly one byte long - in other words a zero-length netstring. But many
netstring packages blindly accept any number of leading zeroes because they use something
like the tolerant strconv.Atoi() to convert the length. Not so for this package.

If the Decoder fails for some reason, the parser stays in a permanent error state as
resynchronizing to the next netstring after any syntax error is impossible to perform
reliably.

# Assembling Messages

Typical usage of netstrings is to assemble a simple message consisting of a small number
of netstrings. E.g., if an application wants to transmit Age, Country and Name it could be
encoded as these three netstrings:

	"2:21,7:Iceland,5:Bjorn,"

and then written to the transmission channel for decoding at the remote end.

To correctly decode the message, the remote end has to know that there are exactly three
netstrings in the message and it also has to know the correct order of the values - in
this case: Age, Country and Name.

# Netstring messages are brittle

As mentioned above, when exchanging messages, both ends have to agree on the number and
order of netstrings to be able to correctly encode and decode the message. If a message
changes, perhaps because a new value is added, both ends have to be upgraded at the same
time. For simple messages and tightly coupled applications, this brittleness is tolerable,
but for loosely coupled applications and more complex messages, this brittleness is
limiting and unwieldy.

To alleviate this brittleness, this package supports the notion of "keyed" netstrings which
provide much greater flexibility in arranging a message.

# "Keyed" netstrings

"Keyed" netstrings allow a single byte - known as a "key" - to be assocated with each
netstring. The "key" only has meaning to the application as this package merely
facilitates associating a "key" with each netstring. For example a "key" might define how
the application should decode the netstring or it might associate a netstring with an
particular field in a struct. Or it might mean something else entirely!

A "keyed" netstring is a simple convention which is nothing more than a regular netstring
with the first byte being used as the "key" and subsequent bytes representing the
value. For example, the netstring:

	"4:dDog,"

can be interpreted as a "keyed" netstring with a key of 'd' and a value of "Dog".

As it's merely a convention, both encoders and decoders need to agree on whether they are
exchanging standard netstrings or "keyed" netstrings.

The benefit of "keyed" netstrings is that they create a simply typing system such that
netstrings can be associated with particular variables and can be serialized in any order
or even optionally serialized as part of an aggregate message. In short, "keyed"
netstrings are a flexible form of Type-Length-Value encoding.

Using the earlier example of Age, Country and Name, a message with "keyed" netstrings
might look like:

	"3:a21,8:CIceland,6:nBjorn,"

where the key 'a' means Age, 'C' means Country and 'n' means Name.

or possibly:

	"6:nBjorn,3:a21,"

if Country is optional.

Note how "keyed" netstrings no longer need to be serialized in order, nor do they need to
be present if optional as compared to positional netstrings.

Another minor benefit of "keyed" netstrings is the ability to differentiate between zero
length values and NULL. If the "keyed" netstring is present it implies a value; if the
"keyed" netstring is absent it implies a NULL.

"Keyed" netstrings thus allow greater flexibility in message assembly and disassembly as
well as much easier upgrades of messages without having to necessarily synchronize
transmitters and receivers.

To ensure "keyed" netstrings remain as strings, a valid "key" must be in the isalpha()
character set - that is 'a'-'z' and 'A'-'Z'.

This package imposes no limitations on how "keyed" netstrings are used. An application is
free to re-use the same "key" in the same message if it makes sense to do so. Note that this level
of flexibility does not apply to the higher level Marshal() and Unmarshal() functions.

Encoder.Marshal and Decoder.Unmarshal are purposely designed with "keyed" netstrings in
mind as they encode and decode a simple struct into a message with "keyed"
netstrings. There are various rules around how netstring keys are used and what
constitutes a simple struct.

# End of Message Strategies

When designing a message containing multiple netstrings, the question arises as to how to
signify to the remote receiver that they have received all netstrings for that particular
message. One strategy already mentioned is to simply agree on the number of netstrings -
with the obvious brittleness that imposes.

Another strategy is to create an encapsulating netstring which contains all the message's
netstrings thus the receiver accepts a single netstring then decodes it for the actually
payload. Using our earlier example, this is what an encapsulating netstring might look
like:

	"26:3:a21,8:CIceland,6:nBjorn,,"

with the encapsulating netstring being 26 bytes long in which the value contains three
"keyed" netstrings of 'a', 'C' and 'n'.

While this strategy works, one problem is that it requires double handling of each
message.

Yet another strategy is to used "keyed" netstrings and designate a particular key as an
end-of-message sentinel, such as 'z'. Using our previous example message with Age, Country
and Name, the complete message with a trailing end-of-message sentinel of 'z' might look
like:

	"3:a21,8:CIceland,6:nBjorn,1:z,"

or possibly:

	"6:nBjorn,3:a21,1:z,"

such that any application decoding the messsage knows it has a complete message when the
'z' "key" is returned.

Encoder.Marshal and Decoder.Unmarshal use this end-of-message sentinel strategy.

# Examples

The _examples directory contains a number of programs which demonstrate various features
of this package so this section merely contains a few fragments to provide a general idea
of idiomatic use.

This example encodes a message into a bytes.Buffer.

	var buf bytes.Buffer
	enc := netstring.NewEncoder(&buf)
	enc.EncodeInt('a', 21)             // Age
	enc.EncodeString('C', "Iceland")   // Country
	enc.EncodeString('n', "Bjorn")     // Name
	enc.EncodeBytes('z')               // End-of-message sentinel
	fmt.Println(buf.String())          // "3:a21,8:CIceland,6:nBjorn,1:z,"

And this example decodes the same message.

	dec := netstring.NewDecoder(&buf)
	k, v, e := dec.DecodeKeyed()       // k=a, v=21
	k, v, e = dec.DecodeKeyed()        // k=C, v=Iceland
	k, v, e = dec.DecodeKeyed()        // k=n, v=Bjorn
	k, v, e = dec.DecodeKeyed()        // k=z End-Of-Message

A complete implementation of this example is in _examples/compare.go which encodes a
simple message with "keyed" netstrings then decodes the message to ensure that the
reconstructed values match the originals.

The higher level functions Marshal() and Unmarshal() can be used to exchange complete
messages. These function used "keyed" netstrings with an end-of-message sentinel to
package up a complete message from a simple struct.

This example encodes the same message as above using Encoder.Marshal().

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

And this example decodes the same message using Decoder.Unmarshal().

	dec := netstring.NewDecoder(&buf)
	in := &msg{}
	dec.Unmarshal('z', in)

_examples/client.go and _example/server.go show how an Encoder and Decoder can be attached
to a network connection such that all exchanges across the network are performed with
netstrings. These example programs use both the lower level Encode*() and Decode*()
functions as well as the higher level Marshal() and Unmarshal() functions.
*/
package netstring
