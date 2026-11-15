package httpsfv

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestMarshalUnsignedInteger(t *testing.T) {
	t.Parallel()

	const maxInteger = uint64(999999999999999)
	type testCase struct {
		value interface{}
		want  string
		err   error
	}
	cases := []testCase{
		{uint(0), "0", nil},
		{uint8(255), "255", nil},
		{uint16(65535), "65535", nil},
		{uint32(4294967295), "4294967295", nil},
		{uint64(0), "0", nil},
		{maxInteger, "999999999999999", nil},
		{maxInteger + 1, "", ErrNumberOutOfRange},
		{uint64(1<<63 - 1), "", ErrNumberOutOfRange},
		{uint64(1 << 63), "", ErrNumberOutOfRange},
		{^uint64(0) - maxInteger, "", ErrNumberOutOfRange},
		{^uint64(0) - maxInteger + 1, "", ErrNumberOutOfRange},
		{^uint64(0) - 1, "", ErrNumberOutOfRange},
		{^uint64(0), "", ErrNumberOutOfRange},
	}
	if strconv.IntSize == 64 {
		cases = append(cases, testCase{^uint(0), "", ErrNumberOutOfRange})
	} else {
		cases = append(cases, testCase{^uint(0), "4294967295", nil})
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%T/%v", c.value, c.value), func(t *testing.T) {
			got, err := Marshal(NewItem(c.value))
			if !errors.Is(err, c.err) || got != c.want {
				t.Errorf("Marshal() = %q, %v; want %q, %v", got, err, c.want, c.err)
			}
		})
	}
}

func TestMarshalUnsignedIntegerOverflow(t *testing.T) {
	t.Parallel()

	item := NewItem(^uint64(0))
	withParams := NewItem(Token("token"))
	withParams.Params.Add("number", ^uint64(0))
	dictionary := NewDictionary()
	dictionary.Add("number", item)
	innerList := InnerList{Items: []Item{NewItem(42), item}, Params: NewParams()}

	for _, value := range []StructuredFieldValue{item, withParams, List{NewItem(42), item}, dictionary, innerList} {
		t.Run(fmt.Sprintf("%T", value), func(t *testing.T) {
			got, err := Marshal(value)
			if !errors.Is(err, ErrNumberOutOfRange) || got != "" {
				t.Errorf("Marshal() = %q, %v; want empty output and ErrNumberOutOfRange", got, err)
			}
		})
	}
}

func TestMarshalItem(t *testing.T) {
	t.Parallel()

	data := []struct {
		in       Item
		expected string
		valid    bool
	}{
		{NewItem(0), "0", true},
		{NewItem(int8(-42)), "-42", true},
		{NewItem(int16(-42)), "-42", true},
		{NewItem(int32(-42)), "-42", true},
		{NewItem(int64(-42)), "-42", true},
		{NewItem(uint(42)), "42", true},
		{NewItem(uint8(42)), "42", true},
		{NewItem(uint16(42)), "42", true},
		{NewItem(uint32(42)), "42", true},
		{NewItem(uint64(42)), "42", true},
		{NewItem(1.1), "1.1", true},
		{NewItem(""), `""`, true},
		{NewItem(Token("foo")), "foo", true},
		{NewItem([]byte{0, 1}), ":AAE=:", true},
		{NewItem(false), "?0", true},
		{NewItem(int64(9999999999999999)), "", false},
		{NewItem(9999999999999999.22), "", false},
		{NewItem("Kévin"), "", false},
		{NewItem(Token("/foo")), "", false},
		{Item{}, "", false},
	}

	for _, d := range data {
		r, err := Marshal(d.in)
		if d.valid && err != nil {
			t.Errorf("error not expected for %v, got %v", d.in, err)
		} else if !d.valid && err == nil {
			t.Errorf("error expected for %v, got %v", d.in, err)
		}

		if r != d.expected {
			t.Errorf("got %v; want %v", r, d.expected)
		}
	}
}

func TestParseItemParamsMarshalSFV(t *testing.T) {
	t.Parallel()

	i := NewItem(Token("bar"))
	i.Params.Add("foo", 0.0)
	i.Params.Add("baz", true)

	var b strings.Builder
	_ = i.marshalSFV(&b)

	if b.String() != "bar;foo=0.0;baz" {
		t.Error("marshalSFV(): invalid")
	}
}

func TestUnmarshalItem(t *testing.T) {
	t.Parallel()

	i1 := NewItem(true)
	i1.Params.Add("foo", true)
	i1.Params.Add("*bar", Token("tok"))

	data := []struct {
		in       []string
		expected Item
		valid    bool
	}{
		{[]string{"?1;foo;*bar=tok"}, i1, false},
		{[]string{"  ?1;foo;*bar=tok  "}, i1, false},
		{[]string{`"foo`, `bar"`}, NewItem("foo, bar"), false},
		{[]string{"é", ""}, Item{}, true},
		{[]string{"tok;é"}, Item{}, true},
		{[]string{"  ?1;foo;*bar=tok  é"}, Item{}, true},
	}

	for _, d := range data {
		i, err := UnmarshalItem(d.in)
		if d.valid && err == nil {
			t.Errorf("UnmarshalItem(%s): error expected", d.in)
		}

		if !d.valid && !reflect.DeepEqual(d.expected, i) {
			t.Errorf("UnmarshalItem(%s) = %v, %v; %v, <nil> expected", d.in, i, err, d.expected)
		}
	}
}

func FuzzUnmarshalItem(f *testing.F) {
	testCases := []string{
		"",
		`"foo";bar;baz=tok`,
		"?0",
		":AAE=:",
		"1.9",
		"-42",
		`%""`,
		`%"K%c3%a9vin"`,
		"@1659578233",
		"@",
		"é",
		"tok;é",
	}

	for _, t := range testCases {
		f.Add(t)
	}

	f.Fuzz(func(t *testing.T, b string) {
		unmarshaled, err := UnmarshalItem([]string{b})
		if err != nil {
			return
		}

		reMarshaled, err := Marshal(unmarshaled)
		if err != nil {
			t.Errorf("Unexpected marshaling error %q for %q, %#v", err, b, unmarshaled)
		}

		reUnmarshaled, err := UnmarshalItem([]string{reMarshaled})
		if err != nil {
			t.Errorf("Unexpected remarshaling error %q for %q; original %q", err, reMarshaled, b)
		}

		if !reflect.DeepEqual(unmarshaled, reUnmarshaled) {
			t.Errorf("Unmarshaled and re-unmarshaled doesn't match: %#v; %#v", unmarshaled, reUnmarshaled)
		}
	})
}
