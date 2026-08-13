package webhook

import (
	"net/http"
	"time"
)

// Verify verifies current-time delivery headers and decodes the event.
func Verify(secret string, headers http.Header, body []byte) (*Delivery, error) {
	return Verifier{Secret: secret}.Verify(headers, body)
}

// VerifyAt verifies a delivery using an explicit clock and tolerance.
func VerifyAt(secret string, headers http.Header, body []byte, now time.Time, tolerance time.Duration) (*Delivery, error) {
	return Verifier{Secret: secret, Tolerance: tolerance, Now: func() time.Time { return now }}.Verify(headers, body)
}
