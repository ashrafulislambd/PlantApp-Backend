package groq

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/chat"
)

func newTestProvider(url string) *ChatProvider {
	p := NewChatProvider("test-key", "")
	p.endpoint = url
	return p
}

func TestReplyStream_ParsesDeltasUntilDone(t *testing.T) {
	var gotReq chatCompletionRequest
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewDecoder(r.Body).Decode(&gotReq)
		w.Header().Set("Content-Type", "text/event-stream")
		for _, line := range []string{
			`data: {"choices":[{"delta":{"role":"assistant","content":""}}]}`,
			`data: {"choices":[{"delta":{"content":"Water "}}]}`,
			`data: {"choices":[{"delta":{"content":"less."}}]}`,
			`data: {"choices":[{"delta":{}}]}`,
			`data: [DONE]`,
		} {
			_, _ = w.Write([]byte(line + "\n\n"))
		}
	}))
	defer srv.Close()

	var pieces []string
	res, err := newTestProvider(srv.URL).ReplyStream(context.Background(),
		[]*chat.Message{{Role: chat.RoleUser, Content: "earlier"}}, "now", "en",
		func(s string) error { pieces = append(pieces, s); return nil })
	if err != nil {
		t.Fatalf("ReplyStream() error = %v", err)
	}
	if strings.Join(pieces, "|") != "Water |less." || res.Text != "Water less." {
		t.Errorf("pieces = %q, text = %q", pieces, res.Text)
	}
	if !gotReq.Stream || len(gotReq.Messages) != 3 || gotReq.Messages[2].Content != "now" {
		t.Errorf("request = %+v; want stream:true and system+history+user", gotReq)
	}
	if gotAuth != "Bearer test-key" {
		t.Errorf("Authorization = %q", gotAuth)
	}
}

func TestReplyStream_MapsStatusErrors(t *testing.T) {
	for _, tc := range []struct {
		status int
		want   error
	}{{http.StatusTooManyRequests, apperr.ErrRateLimited}, {http.StatusInternalServerError, nil}} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tc.status)
		}))
		_, err := newTestProvider(srv.URL).ReplyStream(context.Background(), nil, "x", "en",
			func(string) error { return nil })
		srv.Close()
		if err == nil {
			t.Fatalf("status %d: error = nil", tc.status)
		}
		if tc.want != nil && !errors.Is(err, tc.want) {
			t.Errorf("status %d: error = %v, want %v", tc.status, err, tc.want)
		}
	}
}

func TestReplyStream_EmptyStreamIsAnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer srv.Close()
	_, err := newTestProvider(srv.URL).ReplyStream(context.Background(), nil, "x", "en",
		func(string) error { return nil })
	if err == nil {
		t.Fatal("error = nil, want error for an empty reply")
	}
}

func TestReplyStream_StopsWhenCallbackFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"a\"}}]}\n\ndata: {\"choices\":[{\"delta\":{\"content\":\"b\"}}]}\n\n"))
	}))
	defer srv.Close()
	stop := errors.New("client gone")
	res, err := newTestProvider(srv.URL).ReplyStream(context.Background(), nil, "x", "en",
		func(string) error { return stop })
	if !errors.Is(err, stop) || res.Text != "a" {
		t.Errorf("res = %+v, err = %v; want partial %q and %v", res, err, "a", stop)
	}
}
