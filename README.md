# dextri-pay-go

Go SDK for the Dextri Pay partner API. The module path is temporary until the
public repository is chosen.

For local integration before publication, add the standalone repository with a
Go workspace or a temporary `replace` directive:

```go
require dextri-pay-go v0.0.0

replace dextri-pay-go => ../dextri-pay-go
```

After the permanent repository is chosen, the module path will be changed once
and released as `v1.0.0`; partner projects will then install that public module
with `go get`. Do not publish a version while the temporary module path is in
use.

```go
client, err := dextripay.NewClient(
    "https://pay-api.dextri.com",
    os.Getenv("DEXTRI_PAY_APP_ID"),
    os.Getenv("DEXTRI_PAY_KEY_ID"),
    os.Getenv("DEXTRI_PAY_APP_SECRET"),
)
if err != nil {
    log.Fatal(err)
}

session, err := client.Checkout.CreateDeposit(ctx, dextripay.CreateCheckoutRequest{
    ExternalUserID: "user_1001",
    SourceAsset:    "USDC",
    TargetAsset:    "USDC",
    Amount:         "100.00",
    ReturnURL:      "https://app.example.com/pay/result",
})
```

The SDK signs partner requests, creates idempotency keys, exposes decimal
amounts as strings, maps API errors and verifies webhook signatures. Wallet
connection, QR rendering, user private keys and chain RPC access remain outside
the SDK.

Conversion markets are configured by Admin and identified by `market_id`.
Their denoms and decimal precision come from the active Dextri chain asset
registry; SDK callers use `sell_base` or `buy_base` and never supply precision.
