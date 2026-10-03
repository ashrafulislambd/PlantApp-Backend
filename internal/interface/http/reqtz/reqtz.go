// Package reqtz resolves the requesting user's time zone. Like reqlocale it
// is a leaf package, so handlers can use it without import cycles.
package reqtz

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	// HeaderZone carries an IANA name such as "Asia/Dhaka".
	HeaderZone = "X-Timezone"
	// HeaderOffset carries the UTC offset in minutes (+360 for UTC+6). The
	// mobile app sends this one: Dart only knows the offset, not the IANA name.
	HeaderOffset = "X-UTC-Offset-Minutes"

	maxOffsetMinutes = 14 * 60
)

// Location builds a time zone from an IANA name and/or a UTC offset in
// minutes. A valid name wins; otherwise a valid offset gives a fixed zone;
// otherwise UTC.
func Location(name, offsetMinutes string) *time.Location {
	if name = strings.TrimSpace(name); name != "" && len(name) <= 64 {
		if loc, err := time.LoadLocation(name); err == nil {
			return loc
		}
	}
	if m, err := strconv.Atoi(strings.TrimSpace(offsetMinutes)); err == nil &&
		m >= -maxOffsetMinutes && m <= maxOffsetMinutes {
		return time.FixedZone("UTC"+offsetLabel(m), m*60)
	}
	return time.UTC
}

func offsetLabel(m int) string {
	sign := "+"
	if m < 0 {
		sign, m = "-", -m
	}
	return sign + two(m/60) + ":" + two(m%60)
}

func two(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}

// Resolve returns the request's time zone (UTC when it sent none).
func Resolve(r *http.Request) *time.Location {
	return Location(r.Header.Get(HeaderZone), r.Header.Get(HeaderOffset))
}
