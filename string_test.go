package httpsfv

import (
	"strings"
	"testing"
	"unicode"
)

func TestMarshalString(t *testing.T) {
	t.Parallel()

	data := []struct {
		in       string
		expected string
		valid    bool
	}{
		{"foo", `"foo"`, true},
		{`f"oo`, `"f\"oo"`, true},
		{`f\oo`, `"f\\oo"`, true},
		{`f\"oo`, `"f\\\"oo"`, true},
		{"", `""`, true},
		{"H3lLo", `"H3lLo"`, true},
		{"hel\tlo", `"hel`, false},
		{"hel\x1flo", `"hel`, false},
		{"hel\x7flo", `"hel`, false},
		{"Kévin", `"K`, false},
		{"\t", `"`, false},
	}

	var b strings.Builder

	for _, d := range data {
		b.Reset()

		err := marshalString(&b, d.in)
		if d.valid && err != nil {
			t.Errorf("error not expected for %v, got %v", d.in, err)
		} else if !d.valid && err == nil {
			t.Errorf("error expected for %v, got %v", d.in, err)
		}

		if b.String() != d.expected {
			t.Errorf("got %v; want %v", b.String(), d.expected)
		}
	}
}

func TestParseString(t *testing.T) {
	t.Parallel()

	data := []struct {
		in  string
		out string
		err bool
	}{
		{`"foo"`, "foo", false},
		{`"b\"a\\r"`, `b"a\r`, false},
		{"", "", true},
		{"a", "", true},
		{`"\`, "", true},
		{`"\o`, "", true},
		{string([]byte{'"', 0}), "", true},
		{string([]byte{'"', unicode.MaxASCII}), "", true},
		{`"foo`, "", true},
	}

	for _, d := range data {
		s := &scanner{data: d.in}

		i, err := parseString(s)
		if d.err && err == nil {
			t.Errorf("parseString(%s): error expected", d.in)
		}

		if !d.err && d.out != i {
			t.Errorf("parseString(%s) = %v, %v; %v, <nil> expected", d.in, i, err, d.out)
		}
	}
}
