package webhook

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestVerifyAndDecode(t *testing.T) {
	now := time.UnixMilli(1700000000000)
	cases := []struct {
		name string
		body string
	}{
		{name: "current contract", body: `{"id":"evt_1","type":"operation.succeeded","created_at":"2023-11-14T22:13:20Z","data":{"operation":{"operation_id":"pop_1","external_user_id":"u","type":"deposit","status":"succeeded","source_asset":"USDT","target_asset":"USDC","input_amount":"10.5","output_amount":"10.4","created_at":"2023-11-14T22:13:20Z","updated_at":"2023-11-14T22:13:20Z"}}}`},
		{name: "legacy outbox contract", body: `{"id":"evt_1","type":"operation.succeeded","created_at":"2023-11-14T22:13:20Z","data":{"operation":{"operation_id":"pop_1","external_user_id":"u","type":"deposit","status":"succeeded","source_asset":"USDT","destination_asset":"USDC","amount":"10.5","output_amount":"10.4","created_at":"2023-11-14T22:13:20Z","updated_at":"2023-11-14T22:13:20Z"}}}`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			body := []byte(test.body)
			headers := http.Header{}
			headers.Set(HeaderDeliveryID, "whd_1")
			headers.Set(HeaderTimestamp, "1700000000000")
			headers.Set(HeaderSignature, sign("secret", "whd_1", "1700000000000", body))
			delivery, err := (Verifier{Secret: "secret", Now: func() time.Time { return now }}).Verify(headers, body)
			if err != nil {
				t.Fatal(err)
			}
			operation := delivery.Event.Data.Operation
			if operation.OperationID != "pop_1" || operation.TargetAsset != "USDC" || operation.InputAmount != "10.5" {
				t.Fatalf("operation=%#v", operation)
			}
		})
	}

	body := []byte(cases[0].body)
	headers := http.Header{}
	headers.Set(HeaderDeliveryID, "whd_1")
	headers.Set(HeaderTimestamp, "1700000000000")
	headers.Set(HeaderSignature, sign("secret", "whd_1", "1700000000000", body))
	if _, err := (Verifier{Secret: "secret", Now: func() time.Time { return now }}).Verify(headers, append(body, ' ')); err == nil {
		t.Fatal("tamper accepted")
	}
}

func TestVerifyRejectsInvalidDeliveries(t *testing.T) {
	now := time.UnixMilli(1700000000000)
	validBody := []byte(`{"id":"evt_1","type":"operation.succeeded","created_at":"2023-11-14T22:13:20Z","data":{}}`)
	validHeaders := signedHeaders("secret", "whd_1", "1700000000000", validBody)

	tests := []struct {
		name      string
		verifier  Verifier
		headers   http.Header
		body      []byte
		wantError string
	}{
		{name: "missing secret", verifier: Verifier{Now: func() time.Time { return now }}, headers: validHeaders, body: validBody, wantError: "secret"},
		{name: "missing headers", verifier: Verifier{Secret: "secret", Now: func() time.Time { return now }}, headers: http.Header{}, body: validBody, wantError: "incomplete"},
		{name: "invalid timestamp", verifier: Verifier{Secret: "secret", Now: func() time.Time { return now }}, headers: signedHeaders("secret", "whd_1", "not-a-time", validBody), body: validBody, wantError: "timestamp is invalid"},
		{name: "expired", verifier: Verifier{Secret: "secret", Tolerance: time.Minute, Now: func() time.Time { return now.Add(2 * time.Minute) }}, headers: validHeaders, body: validBody, wantError: "outside"},
		{name: "too far in future", verifier: Verifier{Secret: "secret", Tolerance: time.Minute, Now: func() time.Time { return now.Add(-2 * time.Minute) }}, headers: validHeaders, body: validBody, wantError: "outside"},
		{name: "negative tolerance", verifier: Verifier{Secret: "secret", Tolerance: -1, Now: func() time.Time { return now }}, headers: validHeaders, body: validBody, wantError: "outside"},
		{name: "malformed signature", verifier: Verifier{Secret: "secret", Now: func() time.Time { return now }}, headers: replaceHeader(validHeaders, HeaderSignature, "not-hex"), body: validBody, wantError: "verification failed"},
		{name: "wrong signature", verifier: Verifier{Secret: "secret", Now: func() time.Time { return now }}, headers: signedHeaders("wrong-secret", "whd_1", "1700000000000", validBody), body: validBody, wantError: "verification failed"},
		{name: "invalid JSON", verifier: Verifier{Secret: "secret", Now: func() time.Time { return now }}, headers: signedHeaders("secret", "whd_1", "1700000000000", []byte("{")), body: []byte("{"), wantError: "invalid JSON"},
		{name: "missing event type", verifier: Verifier{Secret: "secret", Now: func() time.Time { return now }}, headers: signedHeaders("secret", "whd_1", "1700000000000", []byte(`{"id":"evt_1"}`)), body: []byte(`{"id":"evt_1"}`), wantError: "missing id or type"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.verifier.Verify(test.headers, test.body)
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("Verify() error = %v, want containing %q", err, test.wantError)
			}
		})
	}
}

func TestVerifyAcceptsTimestampAtToleranceBoundary(t *testing.T) {
	signedAt := time.UnixMilli(1700000000000)
	body := []byte(`{"id":"evt_1","type":"operation.succeeded","created_at":"2023-11-14T22:13:20Z","data":{}}`)
	headers := signedHeaders("secret", "whd_1", "1700000000000", body)
	if _, err := (Verifier{Secret: "secret", Tolerance: time.Minute, Now: func() time.Time { return signedAt.Add(time.Minute) }}).Verify(headers, body); err != nil {
		t.Fatal(err)
	}
}

func FuzzVerifierNeverPanics(f *testing.F) {
	f.Add("delivery", "1700000000000", "signature", "{}")
	f.Fuzz(func(t *testing.T, deliveryID, timestamp, signature, body string) {
		headers := http.Header{}
		headers.Set(HeaderDeliveryID, deliveryID)
		headers.Set(HeaderTimestamp, timestamp)
		headers.Set(HeaderSignature, signature)
		_, _ = (Verifier{Secret: "secret", Now: func() time.Time { return time.UnixMilli(1700000000000) }}).Verify(headers, []byte(body))
	})
}

func signedHeaders(secret, deliveryID, timestamp string, body []byte) http.Header {
	headers := http.Header{}
	headers.Set(HeaderDeliveryID, deliveryID)
	headers.Set(HeaderTimestamp, timestamp)
	headers.Set(HeaderSignature, sign(secret, deliveryID, timestamp, body))
	return headers
}

func replaceHeader(headers http.Header, key, value string) http.Header {
	copy := headers.Clone()
	copy.Set(key, value)
	return copy
}
