package retry

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestDelayUsesAndCapsRetryAfterSeconds(t *testing.T) {
	policy := Policy{BaseDelay: time.Millisecond, MaxDelay: 3 * time.Second, Jitter: func(value time.Duration) time.Duration { return value }}
	if got := policy.Delay(1, "2", time.Unix(0, 0)); got != 2*time.Second {
		t.Fatalf("Retry-After delay = %s, want 2s", got)
	}
	if got := policy.Delay(1, "30", time.Unix(0, 0)); got != 3*time.Second {
		t.Fatalf("capped Retry-After delay = %s, want 3s", got)
	}
}

func TestWaitReturnsPromptlyWhenContextIsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	started := time.Now()
	if err := Wait(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Fatalf("Wait() error = %v, want context.Canceled", err)
	}
	if elapsed := time.Since(started); elapsed > 100*time.Millisecond {
		t.Fatalf("Wait() took %s after context cancellation", elapsed)
	}
}

func TestDelayUsesHTTPDateAndExponentialFallback(t *testing.T) {
	now := time.Date(2026, time.August, 13, 0, 0, 0, 0, time.UTC)
	policy := Policy{BaseDelay: 100 * time.Millisecond, MaxDelay: time.Second, Jitter: func(value time.Duration) time.Duration { return value }}
	if got := policy.Delay(1, now.Add(2*time.Second).Format(http.TimeFormat), now); got != time.Second {
		t.Fatalf("capped HTTP-date delay = %s, want 1s", got)
	}
	if got := policy.Delay(3, "invalid", now); got != 400*time.Millisecond {
		t.Fatalf("fallback delay = %s, want 400ms", got)
	}
}
