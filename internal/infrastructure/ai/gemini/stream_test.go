package gemini

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"plantpal-backend/internal/domain/apperr"
)

func newTestClient(base string) *Client {
	c := New("test-key", "test-model")
	c.baseURL = base
	return c
}

func TestReplyStream_ParsesSSEChunks(t *testing.T) {
	var gotPath, gotQuery, gotKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotQuery, gotKey = r.URL.Path, r.URL.RawQuery, r.Header.Get("x-goog-api-key")
		w.Header().Set("Content-Type", "text/event-stream")
		for _, ev := range []string{
			`data: {"candidates":[{"content":{"role":"model","parts":[{"text":"Mist "}]}}]}`,
			`data: {"candidates":[{"content":{"role":"model","parts":[{"text":"the "},{"text":"leaves."}]}}]}`,
			`data: {"candidates":[{"finishReason":"STOP","content":{"parts":[{"text":""}]}}]}`,
		} {
			_, _ = w.Write([]byte(ev + "\r\n\r\n"))
		}
	}))
	defer srv.Close()

	var pieces []string
	res, err := newTestClient(srv.URL).ReplyStream(context.Background(), nil, "hi", "en",
		func(s string) error { pieces = append(pieces, s); return nil })
	if err != nil {
		t.Fatalf("ReplyStream() error = %v", err)
	}
	if strings.Join(pieces, "|") != "Mist |the leaves." || res.Text != "Mist the leaves." {
		t.Errorf("pieces = %q, text = %q", pieces, res.Text)
	}
	if gotPath != "/test-model:streamGenerateContent" || gotQuery != "alt=sse" || gotKey != "test-key" {
		t.Errorf("path = %q, query = %q, key = %q", gotPath, gotQuery, gotKey)
	}
}

func TestReplyStream_RateLimitAndServerErrors(t *testing.T) {
	for _, status := range []int{http.StatusTooManyRequests, http.StatusBadGateway} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
		}))
		_, err := newTestClient(srv.URL).ReplyStream(context.Background(), nil, "hi", "en",
			func(string) error { return nil })
		srv.Close()
		if err == nil {
			t.Fatalf("status %d: error = nil", status)
		}
		if status == http.StatusTooManyRequests && !errors.Is(err, apperr.ErrRateLimited) {
			t.Errorf("429 error = %v, want ErrRateLimited", err)
		}
	}
}

func TestReplyStream_EmptyReplyIsAnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("data: {\"candidates\":[]}\n\n"))
	}))
	defer srv.Close()
	if _, err := newTestClient(srv.URL).ReplyStream(context.Background(), nil, "hi", "en",
		func(string) error { return nil }); err == nil {
		t.Fatal("error = nil, want error for an empty reply")
	}
}
