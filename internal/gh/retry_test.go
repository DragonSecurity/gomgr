package gh

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
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

func TestJitter_WithinBounds(t *testing.T) {
	const limit = 500 * time.Millisecond
	for i := 0; i < 200; i++ {
		got := jitter(limit)
		if got < 0 || got >= limit {
			t.Fatalf("jitter(%v) = %v, want [0, %v)", limit, got, limit)
		}
	}
}

// Jitter exists to decorrelate concurrent retries, so a generator that
// returned one value would satisfy the bounds check above and still leave
// every caller retrying in step.
func TestJitter_Spreads(t *testing.T) {
	const limit = 500 * time.Millisecond
	seen := make(map[time.Duration]bool)
	for i := 0; i < 100; i++ {
		seen[jitter(limit)] = true
	}
	if len(seen) < 50 {
		t.Errorf("100 draws produced only %d distinct values; jitter is not spreading callers", len(seen))
	}
}

func TestCalcBackoff_GrowsAndCaps(t *testing.T) {
	const maxJitter = 500 * time.Millisecond

	// Each attempt doubles until the cap, and the jitter rides on top, so the
	// bound is the exponential term plus one full jitter width.
	for attempt, want := range map[int]time.Duration{
		0: 500 * time.Millisecond,
		1: 1 * time.Second,
		2: 2 * time.Second,
		3: 4 * time.Second,
	} {
		got := calcBackoff(attempt)
		if got < want || got >= want+maxJitter {
			t.Errorf("calcBackoff(%d) = %v, want [%v, %v)", attempt, got, want, want+maxJitter)
		}
	}

	// The exponential term is clamped so a long retry chain cannot sleep for
	// minutes at a time.
	for _, attempt := range []int{7, 10, 20} {
		got := calcBackoff(attempt)
		if got < 30*time.Second || got >= 30*time.Second+maxJitter {
			t.Errorf("calcBackoff(%d) = %v, want [30s, 30.5s)", attempt, got)
		}
	}
}
