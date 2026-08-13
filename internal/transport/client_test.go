package transport

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/internal/auth"
)

func TestDecodeAPIErrorReadsTopLevelSnakeCaseRequestID(t *testing.T) {
	err := decodeAPIError(http.StatusConflict, "", "idem-1", []byte(`{"code":"IDEMPOTENCY_CONFLICT","message":"request differs","request_id":"req-top","details":{"field":"amount"}}`))
	var apiError *api.APIError
	if !errors.As(err, &apiError) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiError.RequestID != "req-top" || apiError.Code != "IDEMPOTENCY_CONFLICT" || apiError.IdempotencyKey != "idem-1" {
		t.Fatalf("unexpected API error: %#v", apiError)
	}
	var details map[string]string
	if unmarshalErr := json.Unmarshal(apiError.Details, &details); unmarshalErr != nil || details["field"] != "amount" {
		t.Fatalf("unexpected details: %s (%v)", apiError.Details, unmarshalErr)
	}
}

func TestClientDoesNotFollowRedirectsOrLeakSignedHeaders(t *testing.T) {
	requests := 0
	roundTripper := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests++
		if requests > 1 {
			if request.Header.Get("X-DEXTRI-PAY-SIGN") != "" {
				t.Fatal("signed authentication header leaked to redirect target")
			}
			return nil, errors.New("redirect was followed")
		}
		if request.Header.Get("X-DEXTRI-PAY-SIGN") == "" {
			t.Fatal("initial request was not signed")
		}
		return &http.Response{
			StatusCode: http.StatusTemporaryRedirect,
			Header:     http.Header{"Location": []string{"https://target.example/redirected"}},
			Body:       io.NopCloser(strings.NewReader(`{"redirect":"rejected"}`)),
			Request:    request,
		}, nil
	})

	client, err := New(Config{
		BaseURL:     "https://source.example",
		Credentials: auth.Credentials{AppID: "app", KeyID: "key", Secret: "secret"},
		HTTPClient:  &http.Client{Transport: roundTripper},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Do(context.Background(), Request{Method: http.MethodGet, Path: "/redirect"}, nil)
	var apiError *api.APIError
	if !errors.As(err, &apiError) || apiError.StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("expected redirect response as API error, got %v", err)
	}
	if requests != 1 {
		t.Fatalf("SDK made %d requests, want one", requests)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestDecodeAPIErrorReadsNestedEnvelope(t *testing.T) {
	err := decodeAPIError(http.StatusBadRequest, "req-header", "", []byte(`{"error":{"code":"INVALID_AMOUNT","message":"invalid amount","request_id":"req-body"}}`))
	var apiError *api.APIError
	if !errors.As(err, &apiError) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiError.RequestID != "req-header" {
		t.Fatalf("header request ID must win, got %q", apiError.RequestID)
	}
	if apiError.Code != "INVALID_AMOUNT" || apiError.Message != "invalid amount" {
		t.Fatalf("unexpected API error: %#v", apiError)
	}
}

func TestRetryableTransportErrorClassification(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "canceled", err: context.Canceled},
		{name: "DNS not found", err: &net.DNSError{IsNotFound: true}},
		{name: "temporary DNS", err: &net.DNSError{IsTemporary: true}, want: true},
		{name: "DNS timeout", err: &net.DNSError{IsTimeout: true}, want: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := retryableTransportError(test.err); got != test.want {
				t.Fatalf("retryableTransportError() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestClientRejectsOversizedResponse(t *testing.T) {
	roundTripper := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{},
			Body:       io.NopCloser(strings.NewReader(strings.Repeat("x", maxResponseBytes+1))),
			Request:    request,
		}, nil
	})
	client, err := New(Config{
		BaseURL:     "https://source.example",
		Credentials: auth.Credentials{AppID: "app", KeyID: "key", Secret: "secret"},
		HTTPClient:  &http.Client{Transport: roundTripper},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Do(context.Background(), Request{Method: http.MethodPost, Path: "/large"}, &struct{}{})
	var requestError *api.RequestError
	if !errors.As(err, &requestError) || !strings.Contains(requestError.Error(), "exceeds 4 MiB") {
		t.Fatalf("Do() error = %v, want oversized RequestError", err)
	}
}

func TestClientReportsMalformedSuccessJSON(t *testing.T) {
	roundTripper := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		headers := http.Header{}
		headers.Set(auth.HeaderRequestID, "req-json")
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     headers,
			Body:       io.NopCloser(strings.NewReader("{")),
			Request:    request,
		}, nil
	})
	client, err := New(Config{
		BaseURL:     "https://source.example",
		Credentials: auth.Credentials{AppID: "app", KeyID: "key", Secret: "secret"},
		HTTPClient:  &http.Client{Transport: roundTripper},
	})
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.Do(context.Background(), Request{Method: http.MethodPost, Path: "/malformed"}, &struct{}{})
	var requestError *api.RequestError
	if !errors.As(err, &requestError) || response.RequestID != "req-json" || requestError.RequestID != "req-json" {
		t.Fatalf("response=%#v error=%v", response, err)
	}
}
