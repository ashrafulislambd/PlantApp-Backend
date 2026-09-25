// Package groq implements the PriceRefresher port using Groq Compound,
// a model with built-in web-search capability. The key never leaves the
// backend — the Flutter app only calls POST /api/v1/products/{id}/refresh.
package groq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"myplantpal-backend/internal/domain/product"
)

const (
	groqEndpoint = "https://api.groq.com/openai/v1/chat/completions"
	groqModel    = "compound-beta" // Groq Compound: supports real-time web search
	usdToBDT     = 110.0
	cacheTTL     = 6 * time.Hour
)

type cacheEntry struct {
	result *product.RefreshResult
	cachedAt time.Time
}

// PriceRefresher calls Groq Compound with a tightly scoped prompt and
// caches results server-side for 6 hours per product.
type PriceRefresher struct {
	apiKey  string
	client  *http.Client
	cache   map[string]cacheEntry
}

// New creates a PriceRefresher. apiKey is your Groq API key (from env).
func New(apiKey string) *PriceRefresher {
	return &PriceRefresher{
		apiKey: apiKey,
		client: &http.Client{Timeout: 30 * time.Second},
		cache:  make(map[string]cacheEntry),
	}
}

type groqRequest struct {
	Model    string        `json:"model"`
	Messages []groqMessage `json:"messages"`
}

type groqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqResponse struct {
	Choices []struct {
		Message groqMessage `json:"message"`
	} `json:"choices"`
}

type priceJSON struct {
	PriceUSD float64 `json:"price_usd"`
	InStock  bool    `json:"in_stock"`
	Source   string  `json:"source"`
	Note     string  `json:"note"`
}

func (r *PriceRefresher) RefreshPrice(p *product.Product) (*product.RefreshResult, error) {
	// Serve from cache if fresh
	if entry, ok := r.cache[p.ID]; ok && time.Since(entry.cachedAt) < cacheTTL {
		return entry.result, nil
	}

	prompt := fmt.Sprintf(
		`Search the web for the current price of "%s" (vendor: %s) at this URL: %s
		Reply ONLY as valid JSON with these exact keys:
		{"price_usd": <number or null if not found>, "in_stock": <true|false>, "source": "<page title or URL>", "note": "<optional caveat>"}
		Do not add any text outside the JSON object.`,
		p.Name, p.Vendor, p.BuyURL,
	)

	body, _ := json.Marshal(groqRequest{
		Model:    groqModel,
		Messages: []groqMessage{{Role: "user", Content: prompt}},
	})

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, groqEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("groq: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("groq: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("groq: status %d", resp.StatusCode)
	}

	var groqResp groqResponse
	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return nil, fmt.Errorf("groq: decode: %w", err)
	}
	if len(groqResp.Choices) == 0 {
		return nil, fmt.Errorf("groq: no choices returned")
	}

	raw := strings.TrimSpace(groqResp.Choices[0].Message.Content)
	// strip markdown code fences if present
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")

	var pj priceJSON
	if err := json.Unmarshal([]byte(raw), &pj); err != nil {
		return nil, fmt.Errorf("groq: parse price JSON %q: %w", raw, err)
	}

	priceUSD := pj.PriceUSD
	if priceUSD <= 0 {
		priceUSD = p.PriceUSD // fall back to catalog price
	}

	now := time.Now().UTC()
	result := &product.RefreshResult{
		ProductID: p.ID,
		PriceUSD:  priceUSD,
		PriceBDT:  priceUSD * usdToBDT,
		InStock:   pj.InStock,
		AsOf:      now,
		Source:    pj.Source,
		Note:      pj.Note,
	}

	r.cache[p.ID] = cacheEntry{result: result, cachedAt: now}
	return result, nil
}
