package httpsfv

import "testing"

func TestMarshal(t *testing.T) {
	i := NewItem(22.1)
	i.Params.Add("foo", true)
	i.Params.Add("bar", Token("baz"))

	d := NewDictionary()
	d.Add("i", i)

	tok := NewItem(Token("foo"))
	tok.Params.Add("a", "b")
	d.Add("tok", tok)

	res, _ := Marshal(d)
	if res != `i=22.1;foo;bar=baz, tok=foo;a="b"` {
		t.Errorf("marshal: bad result")
	}
}

func TestMarshalError(t *testing.T) {
	if _, err := Marshal(NewItem(Token("à"))); err == nil {
		t.Errorf("marshal: error expected")
	}
}
