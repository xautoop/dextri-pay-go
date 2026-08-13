package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/channels"
	"github.com/xautoop/dextri-pay-go/checkout"
	"github.com/xautoop/dextri-pay-go/conversion"
	"github.com/xautoop/dextri-pay-go/internal/auth"
	"github.com/xautoop/dextri-pay-go/operation"
)

func TestCreateDepositSignsAndRequiresExplicitIdempotency(t *testing.T) {
	fixed := time.UnixMilli(1700000000123)
	var observed api.Response
	client := testClient(t, func(request *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(request.Body)
		if request.URL.Path != "/v1/checkout-sessions" || request.Header.Get(auth.HeaderIdempotency) != "deposit_1" {
			t.Fatalf("bad request %s %q", request.URL.Path, request.Header.Get(auth.HeaderIdempotency))
		}
		var payload map[string]any
		_ = json.Unmarshal(body, &payload)
		if payload["type"] != "deposit" || payload["amount"] != "100.00" {
			t.Fatalf("payload=%#v", payload)
		}
		if request.Header.Get(auth.HeaderTimestamp) != "1700000000123" || request.Header.Get(auth.HeaderNonce) != "nonce_1" {
			t.Fatal("unstable signing inputs")
		}
		return jsonResponse(http.StatusCreated, `{"session_id":"pcs_1","operation_id":"pop_1","expires_at":"2026-08-13T12:00:00Z"}`, http.Header{auth.HeaderRequestID: {"req_1"}}), nil
	}, withClock(func() time.Time { return fixed }), withNonceSource(func() (string, error) { return "nonce_1", nil }), WithResponseObserver(func(response api.Response) { observed = response }))
	session, response, err := client.Checkout.CreateDeposit(context.Background(), checkout.CreateDepositRequest{ExternalUserID: "u1", SourceAsset: "usdc", TargetAsset: "usdc", Amount: "100.00"}, WithIdempotencyKey("deposit_1"))
	if err != nil {
		t.Fatal(err)
	}
	if session.SessionID != "pcs_1" || response.RequestID != "req_1" || observed.RequestID != "req_1" {
		t.Fatalf("unexpected response %#v %#v", session, response)
	}
	if strings.Contains(client.String(), "highly_sensitive") {
		t.Fatal("secret leaked")
	}
}

func TestMutationWithoutIdempotencyDoesNotReachNetwork(t *testing.T) {
	called := false
	client := testClient(t, func(*http.Request) (*http.Response, error) { called = true; return nil, errors.New("unexpected") })
	_, _, err := client.Conversions.CreateQuote(context.Background(), conversion.CreateQuoteRequest{ExternalUserID: "u", MarketID: "asset-a-asset-b", Side: conversion.SideSellBase, InputAmount: "1"})
	if !errors.Is(err, ErrIdempotencyKeyRequired) || called {
		t.Fatalf("err=%v called=%v", err, called)
	}
}

func TestGETRetriesWithFreshNonce(t *testing.T) {
	attempt := 0
	var nonces []string
	client := testClient(t, func(request *http.Request) (*http.Response, error) {
		attempt++
		nonces = append(nonces, request.Header.Get(auth.HeaderNonce))
		if attempt == 1 {
			return jsonResponse(503, `{}`, nil), nil
		}
		return jsonResponse(200, `[]`, nil), nil
	}, WithRetryPolicy(1, time.Nanosecond, time.Nanosecond), withRetryJitter(func(time.Duration) time.Duration { return 0 }), withNonceSource(func() (string, error) { return time.Now().String(), nil }))
	_, response, err := client.Channels.List(context.Background(), channels.ListParams{})
	if err != nil || response.Attempts != 2 || nonces[0] == nonces[1] {
		t.Fatalf("response=%#v err=%v nonces=%#v", response, err, nonces)
	}
}

func TestOperationsListAndAPIError(t *testing.T) {
	client := testClient(t, func(request *http.Request) (*http.Response, error) {
		if request.URL.RawQuery != "external_user_id=user+1&limit=10&status=processing&type=deposit" {
			t.Fatalf("query=%q", request.URL.RawQuery)
		}
		return jsonResponse(409, `{"error":{"code":"IDEMPOTENCY_CONFLICT","message":"different body","request_id":"req_body"}}`, http.Header{auth.HeaderRequestID: {"req_header"}}), nil
	}, WithRetryPolicy(0, time.Millisecond, time.Millisecond))
	_, response, err := client.Operations.List(context.Background(), operation.ListParams{ExternalUserID: "user 1", Type: operation.TypeDeposit, Status: operation.StatusProcessing, Limit: 10})
	var apiError *api.APIError
	if !errors.As(err, &apiError) || !api.IsErrorCode(err, "IDEMPOTENCY_CONFLICT") || response.RequestID != "req_header" {
		t.Fatalf("err=%#v response=%#v", err, response)
	}
}

func TestConfigurationRejectsPlainHTTP(t *testing.T) {
	_, err := New(Config{BaseURL: "http://pay.example", Credentials: Credentials{AppID: "a", KeyID: "k", Secret: "s"}})
	if err == nil {
		t.Fatal("plain HTTP accepted")
	}
}

func testClient(t *testing.T, roundTrip func(*http.Request) (*http.Response, error), options ...Option) *Client {
	t.Helper()
	options = append([]Option{WithHTTPClient(&http.Client{Transport: roundTripFunc(roundTrip)})}, options...)
	client, err := New(Config{BaseURL: "https://pay.test", Credentials: Credentials{AppID: "app", KeyID: "key", Secret: "highly_sensitive"}}, options...)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func jsonResponse(status int, body string, headers http.Header) *http.Response {
	normalized := http.Header{}
	for key, values := range headers {
		for _, value := range values {
			normalized.Add(key, value)
		}
	}
	return &http.Response{StatusCode: status, Header: normalized, Body: io.NopCloser(strings.NewReader(body))}
}
