package gh

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestRetryTransport_SuccessOnFirstAttempt(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{
		Transport: newRetryTransport(http.DefaultTransport, 3),
	}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if c := atomic.LoadInt32(&calls); c != 1 {
		t.Errorf("expected 1 call, got %d", c)
	}
}

func TestRetryTransport_RetriesOn502(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{
		Transport: newRetryTransport(http.DefaultTransport, 3),
	}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 after retries, got %d", resp.StatusCode)
	}
	if c := atomic.LoadInt32(&calls); c != 3 {
		t.Errorf("expected 3 calls, got %d", c)
	}
}

func TestRetryTransport_ExhaustsRetries(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &http.Client{
		Transport: newRetryTransport(http.DefaultTransport, 2),
	}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500 after exhausting retries, got %d", resp.StatusCode)
	}
	// 1 initial + 2 retries = 3 total
	if c := atomic.LoadInt32(&calls); c != 3 {
		t.Errorf("expected 3 calls, got %d", c)
	}
}

func TestRetryTransport_NoRetryOnClientError(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := &http.Client{
		Transport: newRetryTransport(http.DefaultTransport, 3),
	}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if c := atomic.LoadInt32(&calls); c != 1 {
		t.Errorf("expected 1 call (no retry on 404), got %d", c)
	}
}

func TestRetryTransport_RespectsRetryAfter(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{
		Transport: newRetryTransport(http.DefaultTransport, 3),
	}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 after retry, got %d", resp.StatusCode)
	}
	if c := atomic.LoadInt32(&calls); c != 2 {
		t.Errorf("expected 2 calls, got %d", c)
	}
}
