package httpsfv

import (
	"fmt"
	"strings"
	"testing"
)

func TestMarshalDisplayString(t *testing.T) {
	t.Parallel()

	data := []struct {
		in       string
		expected string
		valid    bool
	}{
		{"foo", `%"foo"`, true},
		{"Kévin", `%"K%c3%a9vin"`, true},
	}

	var b strings.Builder

	for _, d := range data {
		b.Reset()

		err := DisplayString(d.in).marshalSFV(&b)
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

func TestParseDisplayString(t *testing.T) {
	t.Parallel()

	data := []struct {
		in  string
		out string
		err bool
	}{
		{`%"foo"`, "foo", false},
		{`%"K%c3%a9vin"`, "Kévin", false},
		{`%"K%00vin"`, "", true},
		{`"K%e9vin"`, "", true},
		{`%K%e9vin"`, "", true},
		{`%"K%e9vin`, "", true},
	}

	for _, d := range data {
		s := &scanner{data: d.in}

		i, err := parseDisplayString(s)
		if d.err && err == nil {
			t.Errorf("parse(%s): error expected", d.in)
		}

		if !d.err && d.out != string(i) {
			fmt.Printf("%q\n", i)
			fmt.Printf("%q\n", d.out)
			t.Errorf("parse(%s) = %v, %v; %v, <nil> expected", d.in, i, err, d.out)
		}
	}
}
