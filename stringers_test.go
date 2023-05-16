package netstring

import (
	"testing"
)

func TestStringers(t *testing.T) {
	s := Key('A').String()
	if s != "A" {
		t.Error("netstring.Key.String() failed", s)
	}

	s = parseFirstByte.String()
	if s != "parseFirstByte" {
		t.Error("netstring.parseState.String() first byte failed", s)
	}

	s = parseComma.String()
	if s != "parseComma" {
		t.Error("netstring.parseState.String() comma failed", s)
	}

	ps := parseState(23)
	s = ps.String()
	if s != "Bizarre parseState" {
		t.Error("netstring.parseState.String() bizarre failed", s)
	}
}
