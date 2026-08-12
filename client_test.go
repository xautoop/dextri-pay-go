package dextripay

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestCreateDepositSignsAndAddsIdempotency(t *testing.T) {
	fixed := time.UnixMilli(1_700_000_000_123)
	var observedMeta ResponseMeta
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		if r.Method != http.MethodPost || r.URL.Path != "/v1/checkout-sessions" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		for header, want := range map[string]string{
			HeaderAppID: "app_test", HeaderKeyID: "key_test", HeaderTimestamp: "1700000000123",
			HeaderNonce: "nonce_test", HeaderContentSHA: ContentSHA256(body),
		} {
			if got := r.Header.Get(header); got != want {
				t.Errorf("%s = %q, want %q", header, got, want)
			}
		}
		if r.Header.Get(HeaderIdempotency) == "" {
			t.Error("missing idempotency key")
		}
		wantSignature := Sign("secret_test", SignInput{
			AppID: "app_test", KeyID: "key_test", TimestampMS: "1700000000123", Nonce: "nonce_test",
			Method: http.MethodPost, CanonicalResource: "/v1/checkout-sessions", ContentSHA256: ContentSHA256(body),
		})
		if r.Header.Get(HeaderSignature) != wantSignature {
			t.Error("signature mismatch")
		}
		var payload map[string]any
		_ = json.Unmarshal(body, &payload)
		if payload["type"] != "deposit" || payload["amount"] != "100.00" {
			t.Errorf("payload = %#v", payload)
		}
		return jsonResponse(http.StatusOK, `{"data":{"session_id":"pcs_1","operation_id":"pop_1","checkout_url":"https://pay.test/c/1","qr_payload":"https://pay.test/c/1","expires_at":"2026-08-12T12:00:00Z"}}`, http.Header{HeaderRequestID: {"req_123"}}), nil
	})}
	client, err := NewClient("https://pay.test", "app_test", "key_test", "secret_test",
		WithHTTPClient(httpClient),
		WithClock(func() time.Time { return fixed }),
		WithNonceSource(func() (string, error) { return "nonce_test", nil }),
		WithResponseObserver(func(meta ResponseMeta) { observedMeta = meta }),
	)
	if err != nil {
		t.Fatal(err)
	}
	session, err := client.Checkout.CreateDeposit(context.Background(), CreateCheckoutRequest{
		ExternalUserID: "user_1", SourceAsset: "USDC", TargetAsset: "USDC", Amount: "100.00",
	})
	if err != nil {
		t.Fatal(err)
	}
	if session.SessionID != "pcs_1" || observedMeta.RequestID != "req_123" {
		t.Fatalf("session/meta = %#v %#v", session, observedMeta)
	}
	if strings.Contains(client.String(), "secret_test") || strings.Contains(client.GoString(), "secret_test") {
		t.Fatal("client formatting leaked secret")
	}
}

func TestGETRetriesWithFreshNonce(t *testing.T) {
	var nonces []string
	attempt := 0
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		nonces = append(nonces, r.Header.Get(HeaderNonce))
		attempt++
		current := attempt
		if current == 1 {
			return jsonResponse(http.StatusServiceUnavailable, `{}`, nil), nil
		}
		return jsonResponse(http.StatusOK, `{"data":[]}`, nil), nil
	})}
	nonceNumber := 0
	client, err := NewClient("https://pay.test", "app", "key", "secret", WithHTTPClient(httpClient), WithMaxRetries(1), WithNonceSource(func() (string, error) {
		nonceNumber++
		return "nonce_" + string(rune('0'+nonceNumber)), nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Channels.List(context.Background(), ListChannelsParams{}); err != nil {
		t.Fatal(err)
	}
	if len(nonces) != 2 || nonces[0] == nonces[1] {
		t.Fatalf("nonces = %#v", nonces)
	}
}

func TestAPIErrorIncludesCodeAndRequestID(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusConflict, `{"error":{"code":"IDEMPOTENCY_CONFLICT","message":"request body differs"}}`, http.Header{HeaderRequestID: {"req_conflict"}}), nil
	})}
	client, err := NewClient("https://pay.test", "app", "key", "secret", WithHTTPClient(httpClient), WithMaxRetries(0))
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Operations.Get(context.Background(), "pop_1")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %T %v", err, err)
	}
	if apiErr.Code != "IDEMPOTENCY_CONFLICT" || apiErr.RequestID != "req_conflict" || apiErr.StatusCode != http.StatusConflict {
		t.Fatalf("api error = %#v", apiErr)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

func jsonResponse(status int, body string, headers http.Header) *http.Response {
	normalized := http.Header{}
	for key, values := range headers {
		for _, value := range values {
			normalized.Add(key, value)
		}
	}
	normalized.Set("Content-Type", "application/json")
	return &http.Response{StatusCode: status, Header: normalized, Body: io.NopCloser(strings.NewReader(body))}
}
