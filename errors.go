package dextripay

import "fmt"

// APIError is a non-2xx response returned by Dextri Pay.
type APIError struct {
	StatusCode int            `json:"-"`
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	RequestID  string         `json:"request_id,omitempty"`
	Details    map[string]any `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	if e.Code == "" {
		return fmt.Sprintf("dextri pay: http %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("dextri pay: %s: %s", e.Code, e.Message)
}

// RequestError wraps a transport or response decoding failure and preserves
// the server request ID when one was available.
type RequestError struct {
	RequestID string
	Err       error
}

func (e *RequestError) Error() string {
	if e.RequestID == "" {
		return fmt.Sprintf("dextri pay request: %v", e.Err)
	}
	return fmt.Sprintf("dextri pay request %s: %v", e.RequestID, e.Err)
}

func (e *RequestError) Unwrap() error { return e.Err }
