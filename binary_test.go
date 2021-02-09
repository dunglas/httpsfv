package httpsfv

import (
	"bytes"
	"strings"
	"testing"
)

func TestBinary(t *testing.T) {
	t.Parallel()

	var bd strings.Builder
	_ = marshalBinary(&bd, []byte{4, 2})

	if bd.String() != ":BAI=:" {
		t.Error("marshalBinary(): invalid")
	}
}

func TestParseBinary(t *testing.T) {
	t.Parallel()

	data := []struct {
		in  string
		out []byte
		err bool
	}{
		{":YWJj:", []byte("abc"), false},
		{":YW55IGNhcm5hbCBwbGVhc3VyZQ==:", []byte("any carnal pleasure"), false},
		{":YW55IGNhcm5hbCBwbGVhc3Vy:", []byte("any carnal pleasur"), false},
		{"", []byte{}, false},
		{":", []byte{}, false},
		{":YW55IGNhcm5hbCBwbGVhc3Vy", []byte{}, false},
		{":YW55IGNhcm5hbCBwbGVhc3Vy~", []byte{}, false},
		{":YW55IGNhcm5hbCBwbGVhc3VyZQ=:", []byte{}, false},
	}

	for _, d := range data {
		s := &scanner{data: d.in}

		i, err := parseBinary(s)
		if d.err && err == nil {
			t.Errorf("parseBinary(%s): error expected", d.in)
		}

		if !d.err && !bytes.Equal(d.out, i) {
			t.Errorf("parseBinary(%s) = %v, %v; %v, <nil> expected", d.in, i, err, d.out)
		}
	}
}
