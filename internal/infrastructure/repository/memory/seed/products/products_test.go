package seed

import (
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// stockImageHosts are not accepted for image URLs: images must be
// vendor-hosted. source.unsplash.com is the retired placeholder host used by
// the original (now-discarded) legacy seed products; kept here as a
// regression guard so it can't be reintroduced by accident.
var stockImageHosts = []string{"unsplash.com", "source.unsplash.com", "pexels.com", "pixabay.com", "wikimedia.org", "alicdn.com"}

const minPerCategory = 2
const minTotalProducts = 50

func TestCategoriesAreValid(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range Categories() {
		if c.ID == "" || strings.TrimSpace(c.Name) == "" {
			t.Errorf("category has empty id or name: %+v", c)
		}
		if c.ID != strings.ToLower(c.ID) {
			t.Errorf("category id %q must be lower-case", c.ID)
		}
		if seen[c.ID] {
			t.Errorf("duplicate category id %q", c.ID)
		}
		seen[c.ID] = true
	}
	if len(seen) != 13 {
		t.Errorf("expected 13 categories, got %d", len(seen))
	}
}

func TestProductsAreValid(t *testing.T) {
	cats := map[string]bool{}
	for _, c := range Categories() {
		cats[c.ID] = true
	}
	ids := map[string]bool{}
	images := map[string]string{}
	for _, p := range Products() {
		if p.ID == "" {
			t.Errorf("product %q has an empty id", p.Name)
			continue
		}
		if ids[p.ID] {
			t.Errorf("duplicate product id %q", p.ID)
		}
		ids[p.ID] = true

		if strings.TrimSpace(p.Name) == "" {
			t.Errorf("%s: empty name", p.ID)
		}
		if !cats[p.CategoryID] {
			t.Errorf("%s: unknown categoryId %q", p.ID, p.CategoryID)
		}
		if strings.TrimSpace(p.Description) == "" {
			t.Errorf("%s: empty description", p.ID)
		}
		if p.PriceBDT <= 0 || p.PriceUSD <= 0 {
			t.Errorf("%s: prices must be positive (bdt=%v usd=%v)", p.ID, p.PriceBDT, p.PriceUSD)
		}
		if strings.TrimSpace(p.Vendor) == "" {
			t.Errorf("%s: empty vendor", p.ID)
		}
		if p.LastVerified.IsZero() {
			t.Errorf("%s: missing or unparsable lastVerified date", p.ID)
		}
		if !strings.HasPrefix(p.ImageURL, "https://") {
			t.Errorf("%s: imageUrl must be an https URL, got %q", p.ID, p.ImageURL)
		}
		if other, dup := images[p.ImageURL]; dup {
			t.Errorf("%s: imageUrl already used by %s", p.ID, other)
		}
		images[p.ImageURL] = p.ID

		for _, host := range stockImageHosts {
			if strings.Contains(p.ImageURL, host) {
				t.Errorf("%s: imageUrl must be vendor-hosted, found %q", p.ID, host)
			}
		}
		if p.BuyURL != "" && !strings.HasPrefix(p.BuyURL, "https://") {
			t.Errorf("%s: buyUrl must be empty or an https URL, got %q", p.ID, p.BuyURL)
		}
	}
}

// TestCatalogCoverage always logs per-category counts (go test -v). Set
// CATALOG_COMPLETE=1 to enforce the final acceptance targets.
func TestCatalogCoverage(t *testing.T) {
	products := Products()
	counts := map[string]int{}
	for _, p := range products {
		counts[p.CategoryID]++
	}
	for _, c := range Categories() {
		t.Logf("%-20s %d products", c.ID, counts[c.ID])
	}
	t.Logf("total products: %d", len(products))

	if os.Getenv("CATALOG_COMPLETE") != "1" {
		return
	}
	if len(products) < minTotalProducts {
		t.Errorf("catalog has %d products, want at least %d", len(products), minTotalProducts)
	}
	for _, c := range Categories() {
		if counts[c.ID] < minPerCategory {
			t.Errorf("category %q has %d products, want at least %d", c.ID, counts[c.ID], minPerCategory)
		}
	}
}

// TestSeedURLsReachable checks every image (and buy) URL over the network. It
// is skipped by default. Run it with:
//
//	CHECK_URLS=1 go test ./internal/infrastructure/repository/memory/seed/products -run Reachable -v
//
// Image failures are errors; buy-URL failures are only logged because shops
// often block automated requests.
func TestSeedURLsReachable(t *testing.T) {
	if os.Getenv("CHECK_URLS") != "1" {
		t.Skip("set CHECK_URLS=1 to check image and buy URLs over the network")
	}
	client := &http.Client{Timeout: 20 * time.Second}
	get := func(url string) (*http.Response, error) {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MyPlantPalSeedCheck/1.0)")
		return client.Do(req)
	}
	for _, p := range Products() {
		resp, err := get(p.ImageURL)
		if err != nil {
			t.Errorf("%s: image request failed: %v", p.ID, err)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK || !strings.HasPrefix(resp.Header.Get("Content-Type"), "image/") {
			t.Errorf("%s: image %s returned %d %q", p.ID, p.ImageURL, resp.StatusCode, resp.Header.Get("Content-Type"))
		}
		if p.BuyURL == "" {
			continue
		}
		resp, err = get(p.BuyURL)
		if err != nil {
			t.Logf("%s: buy URL request failed: %v", p.ID, err)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Logf("%s: buy URL returned %d (may be bot protection)", p.ID, resp.StatusCode)
		}
	}
}
