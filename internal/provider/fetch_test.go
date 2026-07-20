package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"ada-love-ide/internal/config/provider"
)

func TestFetchModels_OpenAICompatible(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			t.Errorf("path = %s, want /models", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer sk-test" {
			t.Errorf("auth header = %q, want Bearer sk-test", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"id":"gpt-4o"},{"id":"text-embedding-3-small"},{"id":"o1-preview"}]}`))
	}))
	defer srv.Close()

	cfg := provider.ProviderConfig{
		APIURL:         srv.URL,
		TypeConnection: "openai",
		APIKeys:        []provider.ProviderAPIKey{{Key: "sk-test"}},
	}

	got, err := FetchModels(context.Background(), cfg)
	if err != nil {
		t.Fatalf("FetchModels err = %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d models, want 3", len(got))
	}
	// First model gpt-4o → vision+tools.
	if got[0].ID != "gpt-4o" || !got[0].Capabilities.Vision || !got[0].Capabilities.Tools {
		t.Errorf("got[0] = %+v", got[0])
	}
	// Embedding model.
	if got[1].ID != "text-embedding-3-small" || !got[1].Capabilities.Embedding {
		t.Errorf("got[1] = %+v", got[1])
	}
}

func TestFetchModels_KeylessType(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if got := r.Header.Get("Authorization"); got != "" {
			t.Errorf("ollama request should not send Authorization, got %q", got)
		}
		w.Write([]byte(`{"data":[{"id":"llama3"}]}`))
	}))
	defer srv.Close()

	cfg := provider.ProviderConfig{
		APIURL:         srv.URL,
		TypeConnection: "ollama",
	}
	got, err := FetchModels(context.Background(), cfg)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if !called {
		t.Error("expected server to be called for keyless type")
	}
	if len(got) != 1 {
		t.Fatalf("got %d models, want 1", len(got))
	}
}

func TestFetchModels_NeedsKeyButNone(t *testing.T) {
	cfg := provider.ProviderConfig{
		APIURL:         "http://example.invalid/v1",
		TypeConnection: "openai",
	}
	_, err := FetchModels(context.Background(), cfg)
	if err != ErrNoAPIKey {
		t.Errorf("err = %v, want ErrNoAPIKey", err)
	}
}

func TestFetchModels_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"bad key"}`))
	}))
	defer srv.Close()

	cfg := provider.ProviderConfig{
		APIURL:         srv.URL,
		TypeConnection: "openai",
		APIKeys:        []provider.ProviderAPIKey{{Key: "k"}},
	}
	_, err := FetchModels(context.Background(), cfg)
	if err == nil {
		t.Error("expected error on 401, got nil")
	}
}
