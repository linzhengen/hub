// Package pagination bounds how much a list endpoint will return.
//
// Every List use case used to apply `limit` only when the caller passed one,
// so the default was "every row". A client that omitted the parameter - which
// the web app did on every screen - made the database materialise the whole
// table and the gateway serialise it. That is fine with a seed database and
// steadily worse with a real one.
//
// The bounds live here rather than in each use case so there is one number to
// change, and so the CLI, the web client and the API reference can all quote
// the same figure.
package pagination

// Bounds on what a list endpoint returns. The values are deliberately modest:
// a caller that genuinely wants more should page.
const (
	// DefaultLimit applies when the caller asks for no particular page size.
	DefaultLimit uint32 = 50
	// MaxLimit caps what a caller may ask for, however large a limit they send.
	MaxLimit uint32 = 200
)

// Page is a limit and offset that have been brought inside the bounds.
//
// It is a distinct type so that a use case cannot accidentally pass the raw,
// unchecked request values to the query builder: the builder takes a Page.
type Page struct {
	limit  uint32
	offset uint32
}

// New clamps a requested page size into the allowed range. A limit of zero
// means "no preference" and gets DefaultLimit, not "unlimited".
func New(limit, offset uint32) Page {
	switch {
	case limit == 0:
		limit = DefaultLimit
	case limit > MaxLimit:
		limit = MaxLimit
	}
	return Page{limit: limit, offset: offset}
}

// Limit is the number of rows to return.
func (p Page) Limit() uint {
	return uint(p.limit)
}

// Offset is the number of rows to skip.
func (p Page) Offset() uint {
	return uint(p.offset)
}
