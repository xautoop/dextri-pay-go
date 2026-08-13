package client

import (
	"errors"
	"strings"
	"sync"
	"testing"
)

func TestResolveIdempotencyKey(t *testing.T) {
	key, err := resolveIdempotencyKey(WithIdempotencyKey(" order-01 "))
	if err != nil || key != "order-01" {
		t.Fatalf("key=%q err=%v", key, err)
	}
	_, err = resolveIdempotencyKey()
	if !errors.Is(err, ErrIdempotencyKeyRequired) {
		t.Fatalf("expected ErrIdempotencyKeyRequired, got %v", err)
	}
	_, err = resolveIdempotencyKey(WithIdempotencyKey(strings.Repeat("x", 129)))
	if err == nil {
		t.Fatal("oversized idempotency key accepted")
	}
	_, err = resolveIdempotencyKey(WithIdempotencyKey("short"))
	if err == nil {
		t.Fatal("short idempotency key accepted")
	}
}

func TestIdempotencyOptionIsSafeForConcurrentReuse(t *testing.T) {
	option := WithIdempotencyKey(" concurrent-order-01 ")
	var group sync.WaitGroup
	for range 32 {
		group.Add(1)
		go func() {
			defer group.Done()
			key, err := resolveIdempotencyKey(option)
			if err != nil || key != "concurrent-order-01" {
				t.Errorf("key=%q err=%v", key, err)
			}
		}()
	}
	group.Wait()
}

func TestNewIdempotencyKey(t *testing.T) {
	first, err := NewIdempotencyKey()
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewIdempotencyKey()
	if err != nil {
		t.Fatal(err)
	}
	if first == second || !strings.HasPrefix(first, "idem_") || len(first) != 37 {
		t.Fatalf("unexpected generated keys %q and %q", first, second)
	}
}
