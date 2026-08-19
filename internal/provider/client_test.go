package provider

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

func testAPIClient(server *httptest.Server, retryMax int) *apiClient {
	client := newRetryableHTTPClient()
	client.HTTPClient = server.Client()
	client.RetryMax = retryMax
	client.RetryWaitMin = time.Millisecond
	client.RetryWaitMax = 5 * time.Millisecond
	client.Backoff = func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
		return 0
	}

	return &apiClient{
		http:      client,
		userAgent: "terraform-provider-redirectpizza-test",
		baseUrl:   server.URL + "/",
		authToken: "test-token",
	}
}

func TestDoRetries429UntilSuccess(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("missing auth header on attempt %d", n)
		}
		if n < 3 {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "Too Many Attempts.", http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	defer server.Close()

	status, body, err := testAPIClient(server, 5).do(context.Background(), http.MethodGet, "v1/redirects/1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != http.StatusOK {
		t.Fatalf("expected status 200, got %d", status)
	}
	if string(body) != `{"ok":true}` {
		t.Fatalf("unexpected body: %s", body)
	}
	if attempts.Load() != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts.Load())
	}
}

func TestDoDoesNotRetryClientErrors(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		http.Error(w, "Resource not found.", http.StatusNotFound)
	}))
	defer server.Close()

	status, body, err := testAPIClient(server, 5).do(context.Background(), http.MethodGet, "v1/redirects/1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", status)
	}
	if attempts.Load() != 1 {
		t.Fatalf("expected a single attempt, got %d", attempts.Load())
	}
	if string(body) == "" {
		t.Fatal("expected response body")
	}
}

func TestDoExhaustsRetriesOnPersistent429(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.Header().Set("Retry-After", "1")
		http.Error(w, "Too Many Attempts.", http.StatusTooManyRequests)
	}))
	defer server.Close()

	_, _, err := testAPIClient(server, 2).do(context.Background(), http.MethodGet, "v1/redirects/1", nil)
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if attempts.Load() != 3 { // initial try + 2 retries
		t.Fatalf("expected 3 attempts, got %d", attempts.Load())
	}
}

func TestRateLimitedBackoffHonorsRetryAfterAndCaps(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header:     make(http.Header),
	}
	resp.Header.Set("Retry-After", "120")

	wait := rateLimitedBackoff(time.Second, 60*time.Second, 1, resp)
	if wait < 60*time.Second || wait > 75*time.Second {
		t.Fatalf("expected capped Retry-After plus jitter around 60s, got %s", wait)
	}
}

func TestRetryableClientRetries429(t *testing.T) {
	retry, err := retryablehttp.DefaultRetryPolicy(context.Background(), &http.Response{StatusCode: http.StatusTooManyRequests}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !retry {
		t.Fatal("expected 429 responses to be retryable")
	}
}
