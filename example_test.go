package httpsfv

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
)

func ExampleUnmarshalList() {
	h := http.Header{}
	h.Add("Preload", `"/member/*/author", "/member/*/comments"`)

	v, err := UnmarshalList(h["Preload"])
	if err != nil {
		log.Fatalln("error: ", err)
	}

	fmt.Println("authors selector: ", v[0].(Item).Value)
	fmt.Println("comments selector: ", v[1].(Item).Value)
	// Output:
	// authors selector:  /member/*/author
	// comments selector:  /member/*/comments
}

func ExampleMarshal() {
	p := List{NewItem("/member/*/author"), NewItem("/member/*/comments")}

	v, err := Marshal(p)
	if err != nil {
		log.Fatalln("error: ", err)
	}

	h := http.Header{}
	h.Set("Preload", v)

	b := new(bytes.Buffer)
	_ = h.Write(b)

	fmt.Println(b.String())
	// Output: Preload: "/member/*/author", "/member/*/comments"
}
