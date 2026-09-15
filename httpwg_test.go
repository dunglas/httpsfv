package httpsfv

import (
	"encoding/base32"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

const (
	ITEM       = "item"
	LIST       = "list"
	DICTIONARY = "dictionary"
)

// test represents a test from the official test suite for the specification.
// See https://github.com/httpwg/structured-field-tests.
type test struct {
	Name       string      `json:"name"`
	Raw        []string    `json:"raw"`
	HeaderType string      `json:"header_type"`
	Expected   interface{} `json:"expected"`
	MustFail   bool        `json:"must_fail"`
	CanFail    bool        `json:"can_fail"`
	Canonical  []string    `json:"canonical"`
}

func valToBareItem(e interface{}) interface{} {
	bareItem, ok := e.(map[string]interface{})
	if !ok {
		if number, ok := e.(json.Number); ok {
			if strings.Contains(number.String(), ".") {
				bi, _ := number.Float64()

				return bi
			}

			bi, _ := number.Int64()

			return bi
		}

		return e
	}

	switch bareItem["__type"] {
	case "binary":
		bi, _ := base32.StdEncoding.DecodeString(bareItem["value"].(string))

		return bi
	case "token":
		return Token(bareItem["value"].(string))
	case "date":
		u, _ := bareItem["value"].(json.Number).Int64()

		return time.Unix(u, 0)
	case "displaystring":
		return DisplayString(bareItem["value"].(string))
	default:
	}

	panic("unknown type " + bareItem["__type"].(string))
}

func populateParams(p *Params, e interface{}) {
	ex := e.([]interface{})
	for _, l := range ex {
		v := l.([]interface{})
		p.Add(v[0].(string), valToBareItem(v[1]))
	}
}

func valToItem(e interface{}) Item {
	if e == nil {
		return Item{}
	}

	ex := e.([]interface{})
	i := NewItem(valToBareItem(ex[0]))
	populateParams(i.Params, ex[1])

	return i
}

func valToInnerList(e []interface{}) InnerList {
	il := InnerList{}
	il.Params = NewParams()

	for _, i := range e[0].([]interface{}) {
		il.Items = append(il.Items, valToItem(i))
	}

	populateParams(il.Params, e[1])

	return il
}

func valToMember(e interface{}) Member {
	il := e.([]interface{})
	if _, ok := il[0].([]interface{}); ok {
		return valToInnerList(il)
	}

	return valToItem(e)
}

func valToList(e interface{}) List {
	if e == nil {
		return nil
	}

	ex := e.([]interface{})
	if len(ex) == 0 {
		return nil
	}

	l := List{}
	for _, m := range ex {
		l = append(l, valToMember(m))
	}

	return l
}

func valToDictionary(e interface{}) *Dictionary {
	if e == nil {
		return nil
	}

	ex := e.([]interface{})
	d := NewDictionary()

	for _, v := range ex {
		m := v.([]interface{})
		d.Add(m[0].(string), valToMember(m[1]))
	}

	return d
}

// listTestFiles returns the JSON test suite files in dir.
func listTestFiles(tb testing.TB, dir string) []string {
	tb.Helper()

	f, err := os.Open(dir)
	if err != nil {
		tb.Fatalf("%s: %s (is the structured-field-tests submodule checked out?)", dir, err)
	}

	defer func() { _ = f.Close() }()

	entries, err := f.Readdir(-1)
	if err != nil {
		tb.Fatalf("%s: %s", dir, err)
	}

	var names []string

	for _, fi := range entries {
		if strings.HasSuffix(fi.Name(), ".json") {
			names = append(names, fi.Name())
		}
	}

	if len(names) == 0 {
		tb.Fatalf("%s: no JSON test file found (is the structured-field-tests submodule checked out?)", dir)
	}

	return names
}

// loadTests decodes the test cases contained in the given test suite file.
func loadTests(tb testing.TB, path string) []test {
	tb.Helper()

	file, err := os.Open(path)
	if err != nil {
		tb.Fatalf("%s: %s (is the structured-field-tests submodule checked out?)", path, err)
	}

	defer func() { _ = file.Close() }()

	dec := json.NewDecoder(file)
	dec.UseNumber()

	var tests []test
	if err := dec.Decode(&tests); err != nil {
		tb.Fatalf("%s: %s", path, err)
	}

	if len(tests) == 0 {
		tb.Fatalf("%s: no test case found", path)
	}

	return tests
}

func TestOfficialTestSuiteParsing(t *testing.T) {
	const dir = "structured-field-tests/"

	for _, n := range listTestFiles(t, dir) {
		for _, te := range loadTests(t, dir+n) {
			t.Run(n+"/"+te.Name, func(t *testing.T) {
				var (
					expected, got StructuredFieldValue
					err           error
				)

				switch te.HeaderType {
				case ITEM:
					expected = valToItem(te.Expected)
					got, err = UnmarshalItem(te.Raw)
				case LIST:
					expected = valToList(te.Expected)
					got, err = UnmarshalList(te.Raw)
				case DICTIONARY:
					expected = valToDictionary(te.Expected)
					got, err = UnmarshalDictionary(te.Raw)
				default:
					panic("unknown header type")
				}

				if te.MustFail && err == nil {
					t.Errorf("%s: %s: must fail", n, te.Name)

					return
				}

				if (!te.MustFail && !te.CanFail) && err != nil {
					t.Errorf("%s: %s: must not fail, got error %s", n, te.Name, err)

					return
				}

				if err == nil && !reflect.DeepEqual(expected, got) {
					t.Errorf("%s: %s: %#v expected, got %#v", n, te.Name, expected, got)
				}
			})
		}
	}
}

func BenchmarkParsingOfficialExamples(b *testing.B) {
	tests := loadTests(b, "structured-field-tests/examples.json")

	for n := 0; n < b.N; n++ {
		for _, te := range tests {
			switch te.HeaderType {
			case ITEM:
				_, _ = UnmarshalItem(te.Raw)
			case LIST:
				_, _ = UnmarshalList(te.Raw)
			case DICTIONARY:
				_, _ = UnmarshalDictionary(te.Raw)
			}
		}
	}
}

func BenchmarkSerializingOfficialExamples(b *testing.B) {
	tests := loadTests(b, "structured-field-tests/examples.json")

	var sfv []StructuredFieldValue

	for _, te := range tests {
		if te.CanFail || te.MustFail {
			continue
		}

		switch te.HeaderType {
		case ITEM:
			sfv = append(sfv, valToItem(te.Expected))
		case LIST:
			sfv = append(sfv, valToList(te.Expected))
		case DICTIONARY:
			sfv = append(sfv, valToDictionary(te.Expected))
		}
	}

	for n := 0; n < b.N; n++ {
		for _, v := range sfv {
			_, _ = Marshal(v)
		}
	}
}

func TestOfficialTestSuiteSerialization(t *testing.T) {
	t.Parallel()

	const dir = "structured-field-tests/serialisation-tests/"

	for _, n := range listTestFiles(t, dir) {
		for _, te := range loadTests(t, dir+n) {
			var sfv StructuredFieldValue

			switch te.HeaderType {
			case ITEM:
				sfv = valToItem(te.Expected)
			case LIST:
				sfv = valToList(te.Expected)
			case DICTIONARY:
				sfv = valToDictionary(te.Expected)
			default:
				panic("unknown header type")
			}

			canonical, err := Marshal(sfv)

			if te.MustFail && err == nil {
				t.Errorf("%s: %s: must fail", n, te.Name)

				continue
			}

			if (!te.MustFail && !te.CanFail) && err != nil {
				t.Errorf("%s: %s: must not fail, got error %s", n, te.Name, err)

				continue
			}

			if err == nil && te.Canonical[0] != canonical {
				t.Errorf("%s: %s: %#v expected, got %#v", n, te.Name, te.Canonical[0], canonical)
			}
		}
	}
}
