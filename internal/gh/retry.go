package gh

import (
	"crypto/rand"
	"math"
	"math/big"
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
	return exp + jitter(500*time.Millisecond)
}

// jitter returns a duration in [0, limit) to spread concurrent retries apart.
//
// crypto/rand is not required here on the merits. Jitter is not a secret and
// an attacker who predicts it gains nothing, which is why this was math/rand
// under a #nosec G404 saying so. It reads from the system source now because
// the explanation cost more than the call does: every scanner flags the weak
// generator, and each one wants that exemption written down, reviewed and
// approved again. A read from the entropy pool on a path that is about to
// sleep for hundreds of milliseconds is cheaper than an exception with a
// maintenance tail behind it.
func jitter(limit time.Duration) time.Duration {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(limit)))
	if err != nil {
		// Not worth abandoning a retry over a failed read from the system
		// entropy source. A constant would be, though: it puts every caller
		// back in step, which is the one thing jitter exists to prevent. The
		// clock is not random, but it is unshared, and spreading callers is
		// all this has to do.
		return time.Duration(time.Now().UnixNano()) % limit
	}
	return time.Duration(n.Int64())
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
