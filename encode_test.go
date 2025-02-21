package httpsfv

import (
	"testing"
	"time"
)

func TestMarshal(t *testing.T) {
	t.Parallel()

	i := NewItem(22.1)
	i.Params.Add("foo", true)
	i.Params.Add("bar", Token("baz"))

	d := NewDictionary()
	d.Add("i", i)

	tok := NewItem(Token("foo"))
	tok.Params.Add("a", "b")
	d.Add("tok", tok)

	date := NewItem(time.Date(1988, 21, 01, 0, 0, 0, 0, time.UTC))
	d.Add("d", date)

	if res, _ := Marshal(d); res != `i=22.1;foo;bar=baz, tok=foo;a="b", d=@620611200` {
		t.Errorf("marshal: bad result: %q", res)
	}
}

func TestMarshalError(t *testing.T) {
	t.Parallel()

	if _, err := Marshal(NewItem(Token("à"))); err == nil {
		t.Errorf("marshal: error expected")
	}
}
