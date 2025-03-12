package httpsfv

import (
	"strings"
	"testing"
)

func TestMarshalBoolean(t *testing.T) {
	t.Parallel()

	var b strings.Builder

	_ = marshalBoolean(&b, true)

	if b.String() != "?1" {
		t.Error("Invalid marshaling")
	}

	b.Reset()
	_ = marshalBoolean(&b, false)

	if b.String() != "?0" {
		t.Error("Invalid marshaling")
	}
}

func TestParseBoolean(t *testing.T) {
	t.Parallel()

	data := []struct {
		in  string
		out bool
		err bool
	}{
		{"?1", true, false},
		{"?0", false, false},
		{"?2", false, true},
		{"", false, true},
		{"?", false, true},
	}

	for _, d := range data {
		s := &scanner{data: d.in}

		i, err := parseBoolean(s)
		if d.err && err == nil {
			t.Errorf("parseBoolean(%s): error expected", d.in)
		}

		if !d.err && d.out != i {
			t.Errorf("parseBoolean(%s) = %v, %v; %v, <nil> expected", d.in, i, err, d.out)
		}
	}
}
