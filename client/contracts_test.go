package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/checkout"
	"github.com/xautoop/dextri-pay-go/conversion"
	"github.com/xautoop/dextri-pay-go/internal/auth"
	"github.com/xautoop/dextri-pay-go/operation"
	"github.com/xautoop/dextri-pay-go/users"
)

func TestCheckoutMutationContracts(t *testing.T) {
	tests := []struct {
		name     string
		wantType checkout.Type
		call     func(*Client) (*checkout.Session, *api.Response, error)
	}{
		{
			name:     "deposit",
			wantType: checkout.TypeDeposit,
			call: func(pay *Client) (*checkout.Session, *api.Response, error) {
				return pay.Checkout.CreateDeposit(context.Background(), checkout.CreateDepositRequest{
					ExternalUserID: "user-01", SourceAsset: "usdt", TargetAsset: "usdc", Amount: "10",
					Metadata: api.Metadata{"order_id": "order-01"},
				}, WithIdempotencyKey("checkout-deposit-01"))
			},
		},
		{
			name:     "withdrawal",
			wantType: checkout.TypeWithdrawal,
			call: func(pay *Client) (*checkout.Session, *api.Response, error) {
				return pay.Checkout.CreateWithdrawal(context.Background(), checkout.CreateWithdrawalRequest{
					ExternalUserID: "user-01", SourceAsset: "usdc", TargetAsset: "usdc", Amount: "10",
				}, WithIdempotencyKey("checkout-withdrawal-01"))
			},
		},
		{
			name:     "conversion",
			wantType: checkout.TypeConversion,
			call: func(pay *Client) (*checkout.Session, *api.Response, error) {
				return pay.Checkout.CreateConversion(context.Background(), checkout.CreateConversionRequest{
					ExternalUserID: "user-01", SourceAsset: "asset-a", TargetAsset: "asset-b", Amount: "10",
				}, WithIdempotencyKey("checkout-conversion-01"))
			},
		},
		{
			name:     "deposit and convert",
			wantType: checkout.TypeDepositAndConvert,
			call: func(pay *Client) (*checkout.Session, *api.Response, error) {
				return pay.Checkout.CreateDepositAndConvert(context.Background(), checkout.CreateDepositAndConvertRequest{
					ExternalUserID: "user-01", SourceAsset: "usdt", TargetAsset: "asset-a", Amount: "10",
				}, WithIdempotencyKey("checkout-deposit-convert-01"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pay := testClient(t, func(request *http.Request) (*http.Response, error) {
				if request.Method != http.MethodPost || request.URL.Path != "/pay/v1/checkout-sessions" {
					t.Fatalf("request=%s %s", request.Method, request.URL.Path)
				}
				if request.Header.Get(auth.HeaderIdempotency) == "" {
					t.Fatal("missing idempotency key")
				}
				var payload struct {
					Type        checkout.Type `json:"type"`
					SourceAsset string        `json:"source_asset"`
					TargetAsset string        `json:"target_asset"`
					Metadata    api.Metadata  `json:"metadata"`
				}
				if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
					t.Fatal(err)
				}
				if payload.Type != test.wantType || payload.SourceAsset == "" || payload.TargetAsset == "" {
					t.Fatalf("payload=%#v", payload)
				}
				if test.wantType == checkout.TypeDeposit && payload.Metadata["order_id"] != "order-01" {
					t.Fatalf("metadata=%#v", payload.Metadata)
				}
				return jsonResponse(http.StatusCreated, `{"session_id":"pcs-01","operation_id":"pop-01","expires_at":"2026-08-13T12:00:00Z"}`, nil), nil
			})
			session, _, err := test.call(pay)
			if err != nil || session.SessionID != "pcs-01" {
				t.Fatalf("session=%#v err=%v", session, err)
			}
		})
	}
}

func TestUserResourceContracts(t *testing.T) {
	pay := testClient(t, func(request *http.Request) (*http.Response, error) {
		switch {
		case request.Method == http.MethodPost && request.URL.Path == "/pay/v1/user-binding-sessions":
			if request.Header.Get(auth.HeaderIdempotency) != "binding-user-01" {
				t.Fatal("binding request missing idempotency key")
			}
			return jsonResponse(http.StatusCreated, `{"session_id":"bind-01","checkout_url":"https://pay.test/checkout","qr_payload":"https://pay.test/checkout","expires_at":"2026-08-13T12:00:00Z"}`, nil), nil
		case request.Method == http.MethodGet && request.URL.EscapedPath() == "/pay/v1/users/user%2F01/balances":
			return jsonResponse(http.StatusOK, `{"external_user_id":"user/01","balances":[{"owner":"dextri1owner","subaccount_id":1,"asset":"asset-a","denom":"uasset-a","available":"9","locked":"1","total":"10"}]}`, nil), nil
		default:
			t.Fatalf("unexpected request %s %s", request.Method, request.URL.EscapedPath())
			return nil, nil
		}
	})

	binding, _, err := pay.Users.CreateBindingSession(context.Background(), users.CreateBindingSessionRequest{
		ExternalUserID: "user/01",
		WalletFamily:   users.WalletFamilyEVM,
	}, WithIdempotencyKey("binding-user-01"))
	if err != nil || binding.SessionID != "bind-01" {
		t.Fatalf("binding=%#v err=%v", binding, err)
	}
	balances, _, err := pay.Users.GetBalances(context.Background(), "user/01")
	if err != nil || len(balances.Items) != 1 || balances.Items[0].Total != "10" {
		t.Fatalf("balances=%#v err=%v", balances, err)
	}
}

func TestConversionResourceContracts(t *testing.T) {
	pay := testClient(t, func(request *http.Request) (*http.Response, error) {
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/pay/v1/conversion-markets":
			return jsonResponse(http.StatusOK, `{"items":[{"app_id":"app","market_id":"asset-a-asset-b","pair":"ASSETA/ASSETB","enabled":true,"buy_base_price":"0.9","sell_base_price":"1.1","updated_at":"2026-08-13T12:00:00Z"}]}`, nil), nil
		case request.Method == http.MethodGet && request.URL.EscapedPath() == "/pay/v1/conversion-markets/asset-a%2Fasset-b":
			return jsonResponse(http.StatusOK, `{"app_id":"app","market_id":"asset-a/asset-b","pair":"ASSETA/ASSETB","enabled":true,"buy_base_price":"0.9","sell_base_price":"1.1","updated_at":"2026-08-13T12:00:00Z"}`, nil), nil
		case request.Method == http.MethodPut && request.URL.Path == "/pay/v1/conversion-markets/asset-a-asset-b/price":
			if request.Header.Get(auth.HeaderIdempotency) != "market-price-01" {
				t.Fatal("price update missing idempotency key")
			}
			return jsonResponse(http.StatusOK, `{"app_id":"app","market_id":"asset-a-asset-b","pair":"ASSETA/ASSETB","enabled":true,"buy_base_price":"0.95","sell_base_price":"1.05","price_version":2,"updated_at":"2026-08-13T12:00:00Z"}`, nil), nil
		case request.Method == http.MethodPost && request.URL.Path == "/pay/v1/conversion-quotes":
			if request.Header.Get(auth.HeaderIdempotency) != "market-quote-01" {
				t.Fatal("quote request missing idempotency key")
			}
			return jsonResponse(http.StatusCreated, `{"quote_id":"quote-01","operation_id":"pop-01","market_id":"asset-a-asset-b","side":"sell_base","input_asset":"asset-a","input_amount":"10","output_asset":"asset-b","output_amount":"9","price":"0.9","expires_at":"2026-08-13T12:00:00Z"}`, nil), nil
		default:
			t.Fatalf("unexpected request %s %s", request.Method, request.URL.EscapedPath())
			return nil, nil
		}
	})

	markets, _, err := pay.Conversions.ListMarkets(context.Background())
	if err != nil || len(markets.Items) != 1 {
		t.Fatalf("markets=%#v err=%v", markets, err)
	}
	market, _, err := pay.Conversions.GetMarket(context.Background(), "asset-a/asset-b")
	if err != nil || market.MarketID != "asset-a/asset-b" {
		t.Fatalf("market=%#v err=%v", market, err)
	}
	market, _, err = pay.Conversions.UpdatePrice(context.Background(), "asset-a-asset-b", conversion.UpdatePriceRequest{
		BuyBasePrice: "0.95", SellBasePrice: "1.05", ExpectedVersion: 1,
	}, WithIdempotencyKey("market-price-01"))
	if err != nil || market.PriceVersion != 2 {
		t.Fatalf("market=%#v err=%v", market, err)
	}
	quote, _, err := pay.Conversions.CreateQuote(context.Background(), conversion.CreateQuoteRequest{
		ExternalUserID: "user-01", MarketID: "asset-a-asset-b", Side: conversion.SideSellBase, InputAmount: "10",
	}, WithIdempotencyKey("market-quote-01"))
	if err != nil || quote.QuoteID != "quote-01" || quote.OutputAmount != "9" {
		t.Fatalf("quote=%#v err=%v", quote, err)
	}
}

func TestInvalidConversionPriceDoesNotReachNetwork(t *testing.T) {
	called := false
	pay := testClient(t, func(*http.Request) (*http.Response, error) {
		called = true
		return nil, nil
	})

	_, _, err := pay.Conversions.UpdatePrice(context.Background(), "asset-a-asset-b", conversion.UpdatePriceRequest{
		BuyBasePrice: "1.1", SellBasePrice: "1.0",
	}, WithIdempotencyKey("invalid-price-01"))
	if err == nil || called {
		t.Fatalf("err=%v called=%v", err, called)
	}
}

func TestOperationGetContract(t *testing.T) {
	pay := testClient(t, func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodGet || request.URL.EscapedPath() != "/pay/v1/operations/pop%2F01" {
			t.Fatalf("request=%s %s", request.Method, request.URL.EscapedPath())
		}
		return jsonResponse(http.StatusOK, `{"operation_id":"pop/01","external_user_id":"user-01","type":"deposit","status":"succeeded","source_asset":"USDT","target_asset":"USDC","input_amount":"10","output_amount":"9.9","metadata":{"order_id":"order-01"},"created_at":"2026-08-13T12:00:00Z","updated_at":"2026-08-13T12:01:00Z"}`, nil), nil
	})

	result, _, err := pay.Operations.Get(context.Background(), "pop/01")
	if err != nil || result.Type != operation.TypeDeposit || result.Metadata["order_id"] != "order-01" {
		t.Fatalf("operation=%#v err=%v", result, err)
	}
}
