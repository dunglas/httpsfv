package httpsfv

import (
	"math"
	"strings"
	"testing"
)

func TestMarshalDecimal(t *testing.T) {
	t.Parallel()

	data := []struct {
		in       float64
		expected string
		valid    bool
	}{
		{10.0, "10.0", true},
		{-10.123, "-10.123", true},
		{10.1236, "10.124", true},
		{-10.0, "-10.0", true},
		{0, "0.0", true},
		{-999999999999.0, "-999999999999.0", true},
		{999999999999.0, "999999999999.0", true},
		{9999999999999, "", false},
		{-9999999999999.0, "", false},
		{9999999999999.0, "", false},
		{1.9, "1.9", true},
		// Ties are decided on the decimal value, not on the binary one.
		{0.5015, "0.502", true},
		{-0.5015, "-0.502", true},
		{0.5025, "0.502", true},
		{2.0005, "2.0", true},
		{1.0025, "1.002", true},
		{0.9995, "1.0", true},
		{999999999999.9995, "", false},
		{-0.0001, "0.0", true},
		{math.NaN(), "", false},
		{math.Inf(1), "", false},
		{math.Inf(-1), "", false},
	}

	var b strings.Builder

	for _, d := range data {
		b.Reset()

		err := marshalDecimal(&b, d.in)
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
