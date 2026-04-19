package embedder

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"socrati/pkg/logger"
)

func newTestLogger() logger.Logger {
	return logger.New("plain", "error")
}

func newClient(t *testing.T, baseURL string, maxRetries int) *VoyageEmbedder {
	t.Helper()
	emb, err := NewVoyage(VoyageConfig{
		APIKey:     "test-key",
		Model:      "voyage-4-lite",
		BaseURL:    baseURL,
		Timeout:    2 * time.Second,
		MaxRetries: maxRetries,
		MaxChars:   8000,
	}, newTestLogger())
	if err != nil {
		t.Fatalf("NewVoyage: %v", err)
	}
	return emb
}

func writeEmbedding(w http.ResponseWriter, vec []float32) {
	_ = json.NewEncoder(w).Encode(voyageResponse{
		Object: "list",
		Data: []struct {
			Object    string    `json:"object"`
			Embedding []float32 `json:"embedding"`
			Index     int       `json:"index"`
		}{{Object: "embedding", Embedding: vec}},
		Model: "voyage-4-lite",
	})
}

func TestEmbed_HappyPath(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("bad Authorization: %q", got)
		}
		if r.URL.Path != "/embeddings" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var req voyageRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("decode req: %v", err)
		}
		if req.Model != "voyage-4-lite" {
			t.Errorf("model=%q", req.Model)
		}
		if len(req.Input) != 1 || req.Input[0] != "hello" {
			t.Errorf("unexpected input %#v", req.Input)
		}
		writeEmbedding(w, []float32{0.1, 0.2, 0.3})
	}))
	t.Cleanup(srv.Close)

	emb := newClient(t, srv.URL, 3)
	vec, err := emb.Embed(context.Background(), "hello")
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}
	if len(vec) != 3 {
		t.Fatalf("dim=%d", len(vec))
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Errorf("calls=%d want 1", atomic.LoadInt32(&calls))
	}
}

func TestEmbed_RetryOn5xxThenSucceed(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			http.Error(w, "boom", http.StatusInternalServerError)
			return
		}
		writeEmbedding(w, []float32{1, 2})
	}))
	t.Cleanup(srv.Close)

	emb := newClient(t, srv.URL, 3)
	vec, err := emb.Embed(context.Background(), "retry me")
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}
	if len(vec) != 2 {
		t.Fatalf("dim=%d", len(vec))
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Errorf("calls=%d want 3", got)
	}
}

func TestEmbed_ExhaustsRetries(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		http.Error(w, "still bad", http.StatusBadGateway)
	}))
	t.Cleanup(srv.Close)

	emb := newClient(t, srv.URL, 2)
	if _, err := emb.Embed(context.Background(), "x"); err == nil {
		t.Fatal("expected error")
	}
	// initial attempt + 2 retries = 3 calls
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Errorf("calls=%d want 3", got)
	}
}

func TestEmbed_NoRetryOn4xx(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		http.Error(w, `{"error":"bad key"}`, http.StatusUnauthorized)
	}))
	t.Cleanup(srv.Close)

	emb := newClient(t, srv.URL, 3)
	if _, err := emb.Embed(context.Background(), "x"); err == nil {
		t.Fatal("expected error")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("calls=%d want 1", got)
	}
}

func TestEmbed_EmptyInputRejected(t *testing.T) {
	emb := newClient(t, "http://unused", 1)
	if _, err := emb.Embed(context.Background(), "   "); err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestEmbed_TruncatesLongInput(t *testing.T) {
	var seen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req voyageRequest
		_ = json.Unmarshal(body, &req)
		if len(req.Input) > 0 {
			seen = req.Input[0]
		}
		writeEmbedding(w, []float32{0})
	}))
	t.Cleanup(srv.Close)

	emb := newClient(t, srv.URL, 1)
	long := strings.Repeat("a", 20000)
	if _, err := emb.Embed(context.Background(), long); err != nil {
		t.Fatalf("Embed: %v", err)
	}
	if len(seen) != 8000 {
		t.Errorf("truncated to %d chars, want 8000", len(seen))
	}
}

func TestNewVoyage_RequiresAPIKey(t *testing.T) {
	if _, err := NewVoyage(VoyageConfig{}, newTestLogger()); err == nil {
		t.Fatal("expected error when api key missing")
	}
}
