# dextri-pay-go

Go SDK for the Dextri Pay partner API. The module path is temporary until the
public repository is chosen.

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
