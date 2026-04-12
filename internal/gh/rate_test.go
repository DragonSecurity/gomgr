package gh

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v84/github"
)

func TestRespectRate_Healthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]any{
			"resources": map[string]any{
				"core": map[string]any{
					"limit":     5000,
					"remaining": 4999,
					"reset":     time.Now().Add(time.Hour).Unix(),
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := github.NewClient(nil)
	url := server.URL + "/"
	client.BaseURL, _ = client.BaseURL.Parse(url)

	err := RespectRate(context.Background(), client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRespectRate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"message": "server error"})
	}))
	defer server.Close()

	client := github.NewClient(nil)
	url := server.URL + "/"
	client.BaseURL, _ = client.BaseURL.Parse(url)

	err := RespectRate(context.Background(), client)
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	if !strings.Contains(err.Error(), "rate limit check") {
		t.Errorf("expected 'rate limit check' in error, got: %v", err)
	}
}
