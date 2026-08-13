package client

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
)

// ErrIdempotencyKeyRequired reports a mutation call without a persisted business key.
var ErrIdempotencyKeyRequired = errors.New("an explicit idempotency key is required for Pay mutation requests")

// NewIdempotencyKey generates a random mutation idempotency key. The caller
// must persist and reuse it for every retry of the same business operation.
func NewIdempotencyKey() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return "idem_" + hex.EncodeToString(raw), nil
}
