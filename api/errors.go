package api

import (
	"encoding/json"
	"errors"
	"fmt"
)

// APIError is a non-2xx response returned by the Pay API.
type APIError struct {
	// StatusCode is the HTTP response status.
	StatusCode int `json:"-"`
	// Code is the stable machine-readable API error code.
	Code string `json:"code"`
	// Message is a safe human-readable error description.
	Message string `json:"message"`
	// RequestID identifies the server request for support and auditing.
	RequestID string `json:"request_id,omitempty"`
	// IdempotencyKey is the mutation key attached by the caller.
	IdempotencyKey string `json:"-"`
	// Details contains forward-compatible structured error context.
	Details json.RawMessage `json:"details,omitempty"`
}

func (err *APIError) Error() string {
	if err == nil {
		return ""
	}
	if err.Code == "" {
		return fmt.Sprintf("dextri pay: http %d: %s", err.StatusCode, err.Message)
	}
	return fmt.Sprintf("dextri pay: %s: %s", err.Code, err.Message)
}

// RequestError reports a transport or response-decoding failure.
type RequestError struct {
	// RequestID identifies the server request when a response was received.
	RequestID string
	// IdempotencyKey is the mutation key attached by the caller.
	IdempotencyKey string
	// Attempts is the total number of HTTP attempts made.
	Attempts int
	// Err is the underlying transport or decoding error.
	Err error
}

func (err *RequestError) Error() string {
	if err == nil {
		return ""
	}
	if err.RequestID == "" {
		return fmt.Sprintf("dextri pay request: %v", err.Err)
	}
	return fmt.Sprintf("dextri pay request %s: %v", err.RequestID, err.Err)
}

func (err *RequestError) Unwrap() error { return err.Err }

// ValidationError reports invalid input before an HTTP request is sent.
type ValidationError struct {
	// Field identifies the invalid request field.
	Field string
	// Message explains the validation requirement.
	Message string
}

func (err *ValidationError) Error() string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("invalid %s: %s", err.Field, err.Message)
}

// IsErrorCode reports whether err contains the supplied stable API error code.
func IsErrorCode(err error, code string) bool {
	var apiError *APIError
	return errors.As(err, &apiError) && apiError.Code == code
}
