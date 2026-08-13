package api

import "net/http"

// Response contains transport metadata for a completed API attempt sequence.
type Response struct {
	// StatusCode is the final HTTP response status.
	StatusCode int
	// RequestID identifies the final server request.
	RequestID string
	// IdempotencyKey is the mutation key attached by the caller.
	IdempotencyKey string
	// Attempts is the total number of HTTP attempts made.
	Attempts int
	// Headers is a clone of the final response headers.
	Headers http.Header
}
