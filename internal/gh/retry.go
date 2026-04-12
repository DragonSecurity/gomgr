package gh

import (
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// retryTransport wraps an http.RoundTripper and retries on transient failures
// (5xx responses and 429 rate limits) with exponential backoff and jitter.
type retryTransport struct {
	base       http.RoundTripper
	maxRetries int
}

// newRetryTransport wraps the given transport with retry logic.
func newRetryTransport(base http.RoundTripper, maxRetries int) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return &retryTransport{base: base, maxRetries: maxRetries}
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= t.maxRetries; attempt++ {
		resp, err = t.base.RoundTrip(req)
		if err != nil {
			// Network-level error: only retry if the request is idempotent or retryable
			if !isRetryableMethod(req.Method) || attempt == t.maxRetries {
				return resp, err
			}
			backoff := calcBackoff(attempt)
			time.Sleep(backoff)
			continue
		}

		if !isRetryableStatus(resp.StatusCode) {
			return resp, nil
		}

		// Don't retry if we've exhausted attempts
		if attempt == t.maxRetries {
			return resp, nil
		}

		// Use Retry-After header if present (GitHub sends it on 429)
		backoff := retryAfterDuration(resp)
		if backoff == 0 {
			backoff = calcBackoff(attempt)
		}

		// Drain and close response body before retry
		_ = resp.Body.Close()
		time.Sleep(backoff)
	}

	return resp, err
}

func isRetryableStatus(status int) bool {
	return status == http.StatusTooManyRequests || // 429
		status == http.StatusInternalServerError || // 500
		status == http.StatusBadGateway || // 502
		status == http.StatusServiceUnavailable || // 503
		status == http.StatusGatewayTimeout // 504
}

func isRetryableMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead
}

// calcBackoff returns exponential backoff with jitter: base * 2^attempt + random jitter.
func calcBackoff(attempt int) time.Duration {
	base := 500 * time.Millisecond
	exp := time.Duration(math.Pow(2, float64(attempt))) * base
	if exp > 30*time.Second {
		exp = 30 * time.Second
	}
	jitter := time.Duration(rand.Int63n(int64(500 * time.Millisecond))) //nolint:gosec
	return exp + jitter
}

// retryAfterDuration parses the Retry-After header if present.
func retryAfterDuration(resp *http.Response) time.Duration {
	ra := resp.Header.Get("Retry-After")
	if ra == "" {
		return 0
	}
	if secs, err := strconv.Atoi(ra); err == nil {
		return time.Duration(secs) * time.Second
	}
	return 0
}
