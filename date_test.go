package httpsfv

import (
	"strings"
	"testing"
	"time"
)

func TestMarshalDate(t *testing.T) {
	t.Parallel()

	data := []struct {
		in       time.Time
		expected string
		valid    bool
	}{
		{time.Unix(1659578233, 0), "@1659578233", true},
		{time.Unix(9999999999999999, 0), "@", false},
	}

	var b strings.Builder

	for _, d := range data {
		b.Reset()

		err := marshalDate(&b, d.in)
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

func TestParseDate(t *testing.T) {
	t.Parallel()

	data := []struct {
		in  string
		out time.Time
		err bool
	}{
		{"@1659578233", time.Unix(1659578233, 0), false},
		{"invalid", time.Time{}, true},
	}

	for _, d := range data {
		s := &scanner{data: d.in}

		i, err := parseDate(s)
		if d.err && err == nil {
			t.Errorf("parse%s): error expected", d.in)
		}

		if !d.err && d.out != i {
			t.Errorf("parse%s) = %v, %v; %v, <nil> expected", d.in, i, err, d.out)
		}
	}
}
