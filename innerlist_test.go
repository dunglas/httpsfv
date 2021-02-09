package httpsfv

import (
	"strings"
	"testing"
)

func TestInnerList(t *testing.T) {
	t.Parallel()

	foo := NewItem("foo")
	foo.Params.Add("a", true)
	foo.Params.Add("b", 1936)

	bar := NewItem(Token("bar"))
	bar.Params.Add("y", []byte{1, 3, 1, 2})

	params := NewParams()
	params.Add("d", 18.71)

	i := InnerList{
		[]Item{foo, bar},
		params,
	}

	var b strings.Builder
	_ = i.marshalSFV(&b)

	if b.String() != `("foo";a;b=1936 bar;y=:AQMBAg==:);d=18.71` {
		t.Errorf("invalid marshalSFV(): %v", b.String())
	}
}
