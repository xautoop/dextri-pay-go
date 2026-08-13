// Package retry implements bounded Retry-After-aware backoff.
package retry

import (
	"context"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Policy struct {
	MaxRetries          int
	BaseDelay, MaxDelay time.Duration
	Jitter              func(time.Duration) time.Duration
}

func (policy Policy) Normalize() Policy {
	if policy.MaxRetries < 0 {
		policy.MaxRetries = 0
	}
	if policy.MaxRetries > 5 {
		policy.MaxRetries = 5
	}
	if policy.BaseDelay <= 0 {
		policy.BaseDelay = 200 * time.Millisecond
	}
	if policy.MaxDelay <= 0 {
		policy.MaxDelay = 5 * time.Second
	}
	if policy.MaxDelay < policy.BaseDelay {
		policy.MaxDelay = policy.BaseDelay
	}
	if policy.Jitter == nil {
		policy.Jitter = func(max time.Duration) time.Duration {
			if max <= 0 {
				return 0
			}
			return time.Duration(rand.Int64N(int64(max) + 1))
		}
	}
	return policy
}

func (policy Policy) Delay(attempt int, retryAfter string, now time.Time) time.Duration {
	policy = policy.Normalize()
	if delay, ok := parseRetryAfter(retryAfter, now); ok {
		if delay > policy.MaxDelay {
			return policy.MaxDelay
		}
		return delay
	}
	if attempt < 1 {
		attempt = 1
	}
	maximum := policy.BaseDelay
	for current := 1; current < attempt && maximum < policy.MaxDelay; current++ {
		maximum *= 2
		if maximum > policy.MaxDelay {
			maximum = policy.MaxDelay
		}
	}
	return policy.Jitter(maximum)
}

func Wait(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func parseRetryAfter(raw string, now time.Time) (time.Duration, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false
	}
	if seconds, err := strconv.ParseInt(raw, 10, 32); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second, true
	}
	parsed, err := http.ParseTime(raw)
	if err != nil {
		return 0, false
	}
	delay := parsed.Sub(now)
	if delay < 0 {
		delay = 0
	}
	return delay, true
}
