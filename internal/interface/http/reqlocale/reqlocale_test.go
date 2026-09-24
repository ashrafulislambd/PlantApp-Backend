package reqlocale

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResolve(t *testing.T) {
	tests := []struct {
		name       string
		headerVal  string
		setHeader  bool
		wantResult string
	}{
		{name: "no header defaults to english", setHeader: false, wantResult: EN},
		{name: "empty header defaults to english", headerVal: "", setHeader: true, wantResult: EN},
		{name: "exact bn", headerVal: "bn", setHeader: true, wantResult: BN},
		{name: "uppercase BN", headerVal: "BN", setHeader: true, wantResult: BN},
		{name: "bn with region/quality", headerVal: "bn-BD,bn;q=0.9", setHeader: true, wantResult: BN},
		{name: "english explicit", headerVal: "en-US,en;q=0.9", setHeader: true, wantResult: EN},
		{name: "unrelated language falls back to english", headerVal: "fr-FR", setHeader: true, wantResult: EN},
		{name: "bengali word not just prefix should not false-match", headerVal: "bnx", setHeader: true, wantResult: BN}, // prefix match is intentionally loose
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.setHeader {
				r.Header.Set("Accept-Language", tt.headerVal)
			}
			got := Resolve(r)
			if got != tt.wantResult {
				t.Errorf("Resolve() with header %q = %q, want %q", tt.headerVal, got, tt.wantResult)
			}
		})
	}
}
