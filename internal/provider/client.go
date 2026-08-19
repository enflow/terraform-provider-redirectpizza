package provider

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	httpRetryMax     = 10
	httpRetryWaitMin = time.Second
	httpRetryWaitMax = 60 * time.Second
)

func newRetryableHTTPClient() *retryablehttp.Client {
	client := retryablehttp.NewClient()
	client.RetryMax = httpRetryMax
	client.RetryWaitMin = httpRetryWaitMin
	client.RetryWaitMax = httpRetryWaitMax
	client.Logger = nil
	client.Backoff = rateLimitedBackoff
	client.RequestLogHook = func(_ retryablehttp.Logger, req *http.Request, retryNumber int) {
		if retryNumber > 0 {
			tflog.Warn(req.Context(), "Retrying redirect.pizza API request", map[string]interface{}{
				"method": req.Method,
				"url":    req.URL.String(),
				"retry":  retryNumber,
			})
		}
	}

	return client
}

// rateLimitedBackoff honors Retry-After (via DefaultBackoff), caps the wait,
// and adds jitter so parallel Terraform workers do not retry in lockstep.
func rateLimitedBackoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	wait := retryablehttp.DefaultBackoff(min, max, attemptNum, resp)
	if wait > max {
		wait = max
	}
	if wait <= 0 {
		wait = min
	}

	jitter := time.Duration(rand.Int63n(int64(wait/4) + 1))

	return wait + jitter
}

func (c *apiClient) do(ctx context.Context, method, path string, body []byte) (int, []byte, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, method, c.baseUrl+path, reader)
	if err != nil {
		return 0, nil, fmt.Errorf("cannot create http request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.authToken)
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("cannot read response body: %w", err)
	}

	return resp.StatusCode, respBody, nil
}
