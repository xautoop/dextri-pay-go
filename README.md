# Dextri Pay Go SDK

English | [简体中文](README.zh-CN.md)

The official Go SDK for the Dextri Pay Partner API. It provides authenticated API access for channels, Hosted Checkout, user balances, App-configured conversions, operations, and webhook verification.

## Status

This repository is currently pre-release and has no published version tag. The first planned release is `v0.1.0`. It requires Go 1.25.4 or later, and public APIs may change before `v1.0.0`.

The SDK is standalone. It does not require a Dextri source checkout, database, blockchain node, wallet library, or any private service package.

## Install

Published versions can be installed with:

```bash
go get github.com/xautoop/dextri-pay-go@latest
```

## Get credentials

Request Dextri Pay access at [veraxon.xyz](https://veraxon.xyz). Credentials are issued after administrator review. When applying, provide:

- application name and owner;
- environment: Sandbox or Production;
- allowed return domains;
- required deposit, withdrawal, and conversion capabilities;
- expected transaction limits;
- webhook URL and optional outbound IP allowlist.

After approval, you receive an API base URL and an App credential containing `app_id`, `key_id`, and `app_secret`. The secret is shown only once. Store it in a server-side secret manager and never expose it to a browser, mobile app, log, or source repository.

Sandbox and Production credentials are isolated and cannot be mixed.

## Create a client

```go
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/xautoop/dextri-pay-go/client"
)

func newPayClient() *client.Client {
	httpClient := &http.Client{Timeout: 15 * time.Second}

	pay, err := client.New(client.Config{
		BaseURL: os.Getenv("DEXTRI_PAY_BASE_URL"),
		Credentials: client.Credentials{
			AppID:  os.Getenv("DEXTRI_PAY_APP_ID"),
			KeyID:  os.Getenv("DEXTRI_PAY_KEY_ID"),
			Secret: os.Getenv("DEXTRI_PAY_APP_SECRET"),
		},
		UserAgent: "merchant-service/1.0.0",
	}, client.WithHTTPClient(httpClient))
	if err != nil {
		log.Fatal(err)
	}
	return pay
}
```

Production endpoints must use HTTPS. Plain HTTP can only be enabled explicitly for a loopback development server with `client.WithAllowInsecureLoopbackHTTP()`.

## Create a deposit Checkout

All mutation requests require an idempotency key. Generate and persist one key for each business operation, and reuse the same key when retrying that operation.

```go
package main

import (
	"context"

	"github.com/xautoop/dextri-pay-go/checkout"
	"github.com/xautoop/dextri-pay-go/client"
	"github.com/xautoop/dextri-pay-go/money"
)

func createDeposit(ctx context.Context, pay *client.Client) (string, error) {
	session, _, err := pay.Checkout.CreateDeposit(ctx, checkout.CreateDepositRequest{
		ExternalUserID:    "user_1001",
		ClientReferenceID: "deposit_20260813_001",
		SourceAsset:       "USDT",
		TargetAsset:       "USDC",
		Amount:            money.Decimal("100.00"),
		ReturnURL:         "https://merchant.example/pay/result",
	}, client.WithIdempotencyKey("deposit_20260813_001"))
	if err != nil {
		return "", err
	}
	return session.CheckoutURL, nil
}
```

Show `CheckoutURL` directly or render `QRPayload` as a QR code. Wallet connection and user authorization happen in Hosted Checkout; the SDK never handles user private keys.

The same client also exposes:

- `pay.Channels.List` for channels currently authorized and available to the App;
- `pay.Checkout.CreateWithdrawal`, `CreateConversion`, and `CreateDepositAndConvert`;
- `pay.Users.CreateBindingSession` and `GetBalances`;
- `pay.Conversions.ListMarkets`, `GetMarket`, `UpdatePrice`, and `CreateQuote`;
- `pay.Operations.Get` and `List`.

Assets, markets, denominations, and precision are supplied by the API according to chain and Admin configuration. The SDK does not hardcode a token list or perform ledger calculations with floating-point values. Monetary values use `money.Decimal`, which is encoded as a JSON string.

## Error handling

```go
var apiErr *api.APIError
if errors.As(err, &apiErr) {
	log.Printf("code=%s request_id=%s", apiErr.Code, apiErr.RequestID)
}

if api.IsErrorCode(err, "IDEMPOTENCY_CONFLICT") {
	// Do not retry with a different payload under the same idempotency key.
}
```

The SDK only retries safe requests or mutation requests that contain an idempotency key, and honors `Retry-After`. Always set a request deadline or use an `http.Client` timeout.

## Verify webhooks

Verify the signature before processing the event, then deduplicate events by `delivery.Event.ID` in your application.

```go
delivery, err := webhook.Verify(webhookSecret, request.Header, body)
if err != nil {
	// Reject the request.
}

event := delivery.Event
```

The webhook secret is separate from the App API secret.

## Packages

- `client`: authenticated client and API services;
- `api`: response metadata, API errors, and common JSON types;
- `channels`, `checkout`, `conversion`, `operation`, `users`: request and response types;
- `money`: decimal-string monetary values;
- `webhook`: webhook signature verification and event types.

See [SDK Architecture](docs/architecture.md) for package boundaries and dependency direction.

## Development

Run the repository checks without any sibling repository:

```bash
make check
```

The check includes formatting, module tidiness, unit tests, the race detector, `go vet`, Staticcheck, and `git diff --check`.

## License

Licensed under the [Apache License 2.0](LICENSE).
