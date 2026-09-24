// Package reqlocale resolves which language a request wants, from the
// standard Accept-Language header. It's a leaf package (no dependency on
// the router) so both the top-level http package and versioned handler
// packages can depend on it without an import cycle.
package reqlocale

import (
	"net/http"
	"strings"
)

const (
	EN = "en"
	BN = "bn"
)

// Resolve returns BN if the request's Accept-Language header indicates
// Bengali, otherwise EN. This app ships exactly two UI languages, so a
// simple prefix check is enough — no need for full BCP-47 negotiation.
func Resolve(r *http.Request) string {
	al := strings.ToLower(r.Header.Get("Accept-Language"))
	if strings.HasPrefix(al, BN) {
		return BN
	}
	return EN
}
