package httpsfv

import (
	"reflect"
	"testing"
)

func TestMarshalList(t *testing.T) {
	t.Parallel()

	params := NewParams()
	params.Add("foo", true)
	params.Add("bar", Token("baz"))

	tokItem := NewItem(Token("tok"))
	tokItem.Params.Add("tp1", 42.42)
	tokItem.Params.Add("tp2", []byte{0, 1})

	il := InnerList{
		[]Item{NewItem("il"), tokItem},
		NewParams(),
	}
	il.Params.Add("ilp1", true)
	il.Params.Add("ilp2", false)

	data := []struct {
		in       List
		expected string
		valid    bool
	}{
		{List{}, "", true},
		{List{NewItem(true)}, "?1", true},
		{List{Item{"hello", params}}, `"hello";foo;bar=baz`, true},
		{List{il, Item{"hi", params}}, `("il" tok;tp1=42.42;tp2=:AAE=:);ilp1;ilp2=?0, "hi";foo;bar=baz`, true},
		{List{NewItem(Token("é"))}, "", false},
		{List{Item{}}, "", false},
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

func TestUnmarshalList(t *testing.T) {
	t.Parallel()

	l1 := List{Item{Token("foo"), NewParams()}, Item{Token("bar"), NewParams()}}

	il2 := Item{"foo", NewParams()}
	l2 := List{il2}
	il2.Params.Add("bar", true)
	il2.Params.Add("baz", Token("tok"))

	il3 := InnerList{[]Item{{Token("foo"), NewParams()}, {Token("bar"), NewParams()}}, NewParams()}
	il3.Params.Add("bat", true)
	l3 := List{il3}

	data := []struct {
		in  []string
		out List
		err bool
	}{
		{[]string{""}, nil, false},
		{[]string{"foo,bar"}, l1, false},
		{[]string{"foo, bar"}, l1, false},
		{[]string{"foo,\t bar"}, l1, false},
		{[]string{"foo", "bar"}, l1, false},
		{[]string{`"foo";bar;baz=tok`}, l2, false},
		{[]string{`(foo bar);bat`}, l3, false},
		{[]string{`()`}, List{InnerList{nil, NewParams()}}, false},
		{[]string{`   "foo";bar;baz=tok,  (foo bar);bat `}, List{il2, il3}, false},
		{[]string{`foo,bar,`}, nil, true},
		{[]string{`foo,baré`}, nil, true},
		{[]string{`é`}, nil, true},
		{[]string{`foo,"bar"  é`}, nil, true},
		{[]string{`(foo `}, nil, true},
		{[]string{`(foo);é`}, nil, true},
		{[]string{`("é")`}, nil, true},
		{[]string{`(""`}, nil, true},
		{[]string{`(`}, nil, true},
	}

	for _, d := range data {
		l, err := UnmarshalList(d.in)
		if d.err && err == nil {
			t.Errorf("UnmarshalList(%s): error expected", d.in)
		}

		if !d.err && !reflect.DeepEqual(d.out, l) {
			t.Errorf("UnmarshalList(%s) = %t, %v; %t, <nil> expected", d.in, l, err, d.out)
		}
	}
}

func FuzzUnmarshalList(f *testing.F) {
	testCases := []string{"",
		"foo,bar",
		"foo, bar",
		"foo,\t bar",
		"foo", "bar",
		`"foo";bar;baz=tok`,
		`(foo bar);bat`,
		`()`,
		`   "foo";bar;baz=tok,  (foo bar);bat `,
		`foo,bar,`,
		`foo,baré`,
		`é`,
		`foo,"bar"  é`,
		`(foo `,
		`(foo);é`,
		`("é")`,
		`(""`,
		`(`,
		"1.9",
	}

	for _, t := range testCases {
		f.Add(t)
	}

	f.Fuzz(func(t *testing.T, b string) {
		unmarshaled, err := UnmarshalList([]string{b})
		if err != nil {
			return
		}

		reMarshaled, err := Marshal(unmarshaled)
		if err != nil {
			t.Errorf("Unexpected marshaling error %q for %q, %#v", err, b, unmarshaled)
		}

		reUnmarshaled, err := UnmarshalList([]string{reMarshaled})
		if err != nil {
			t.Errorf("Unexpected remarshaling error %q for %q; original %q", err, reMarshaled, b)
		}

		if !reflect.DeepEqual(unmarshaled, reUnmarshaled) {
			t.Errorf("Unmarshaled and re-unmarshaled doesn't match: %#v; %#v", unmarshaled, reUnmarshaled)
		}
	})
}
