package client

import (
	"errors"
	"net/http"
	"time"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/internal/retry"
)

// Option configures client transport behavior without changing API DTOs.
type Option func(*options) error

type options struct {
	httpClient        *http.Client
	retry             retry.Policy
	allowInsecureHTTP bool
	observer          func(api.Response)
	now               func() time.Time
	nonce             func() (string, error)
}

func defaultOptions() options {
	return options{retry: retry.Policy{MaxRetries: 2, BaseDelay: 200 * time.Millisecond, MaxDelay: 5 * time.Second}}
}

// WithHTTPClient uses the supplied HTTP client for all API requests.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(options *options) error {
		if httpClient == nil {
			return errors.New("http client is nil")
		}
		options.httpClient = httpClient
		return nil
	}
}

// WithRetryPolicy configures bounded retries for safe or idempotent requests.
func WithRetryPolicy(maxRetries int, baseDelay, maxDelay time.Duration) Option {
	return func(options *options) error {
		if maxRetries < 0 || maxRetries > 5 {
			return errors.New("max retries must be between 0 and 5")
		}
		if baseDelay <= 0 || maxDelay < baseDelay {
			return errors.New("retry delays are invalid")
		}
		options.retry.MaxRetries, options.retry.BaseDelay, options.retry.MaxDelay = maxRetries, baseDelay, maxDelay
		return nil
	}
}

// WithAllowInsecureLoopbackHTTP permits explicit HTTP only for local loopback development.
func WithAllowInsecureLoopbackHTTP() Option {
	return func(options *options) error { options.allowInsecureHTTP = true; return nil }
}

// WithResponseObserver receives final response metadata for every completed HTTP response.
func WithResponseObserver(observer func(api.Response)) Option {
	return func(options *options) error {
		if observer == nil {
			return errors.New("response observer is nil")
		}
		options.observer = observer
		return nil
	}
}

func withClock(clock func() time.Time) Option {
	return func(options *options) error { options.now = clock; return nil }
}

func withNonceSource(source func() (string, error)) Option {
	return func(options *options) error { options.nonce = source; return nil }
}

func withRetryJitter(jitter func(time.Duration) time.Duration) Option {
	return func(options *options) error { options.retry.Jitter = jitter; return nil }
}
