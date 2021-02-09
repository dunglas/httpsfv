package httpsfv

// Member is a marker interface for members of dictionaries and lists.
//
// See https://httpwg.org/specs/rfc8941.html#list.
type Member interface {
	member()
	marshaler
}
