package reqtz

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestResolve(t *testing.T) {
	tests := []struct {
		name       string
		zone, off  string
		wantOffset int // seconds east of UTC at the reference instant
	}{
		{"nothing sent is UTC", "", "", 0},
		{"offset in minutes", "", "360", 6 * 3600},
		{"negative offset", "", "-300", -5 * 3600},
		{"IANA name wins over offset", "Asia/Dhaka", "0", 6 * 3600},
		{"bad name falls back to offset", "Not/AZone", "330", 5*3600 + 1800},
		{"garbage offset is UTC", "", "abc", 0},
		{"absurd offset is UTC", "", "5000", 0},
	}
	ref := time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.zone != "" {
				r.Header.Set(HeaderZone, tt.zone)
			}
			if tt.off != "" {
				r.Header.Set(HeaderOffset, tt.off)
			}
			_, off := ref.In(Resolve(r)).Zone()
			if off != tt.wantOffset {
				t.Errorf("offset = %d, want %d", off, tt.wantOffset)
			}
		})
	}
}
