package client

import (
	"errors"
	"strings"
)

// RequestOption configures one Partner API request.
type RequestOption func(*requestOptions) error

type requestOptions struct {
	idempotencyKey string
}

// WithIdempotencyKey attaches a caller-persisted key to a mutation request.
func WithIdempotencyKey(key string) RequestOption {
	normalized := strings.TrimSpace(key)
	var validationErr error
	if len(normalized) < 8 || len(normalized) > 128 {
		validationErr = errors.New("idempotency key must contain 8 to 128 characters")
	}
	return func(options *requestOptions) error {
		if validationErr != nil {
			return validationErr
		}
		options.idempotencyKey = normalized
		return nil
	}
}

func resolveIdempotencyKey(supplied ...RequestOption) (string, error) {
	var options requestOptions
	for _, option := range supplied {
		if option != nil {
			if err := option(&options); err != nil {
				return "", err
			}
		}
	}
	if options.idempotencyKey == "" {
		return "", ErrIdempotencyKeyRequired
	}
	return options.idempotencyKey, nil
}
