package httpsfv

// Member is a marker interface for members of dictionaries and lists.
//
// See https://httpwg.org/http-extensions/draft-ietf-httpbis-header-structure.html#list.
type Member interface {
	member()
	marshaler
}
