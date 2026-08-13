package main

import (
	"context"
	"fmt"
	"os"

	"github.com/xautoop/dextri-pay-go/checkout"
	"github.com/xautoop/dextri-pay-go/client"
	"github.com/xautoop/dextri-pay-go/money"
)

func main() {
	pay, err := client.New(client.Config{
		BaseURL: os.Getenv("DEXTRI_PAY_BASE_URL"),
		Credentials: client.Credentials{
			AppID:  os.Getenv("DEXTRI_PAY_APP_ID"),
			KeyID:  os.Getenv("DEXTRI_PAY_KEY_ID"),
			Secret: os.Getenv("DEXTRI_PAY_APP_SECRET"),
		},
		UserAgent: "dextri-pay-deposit-example/1.0.0",
	})
	if err != nil {
		panic(err)
	}
	session, _, err := pay.Checkout.CreateDeposit(context.Background(), checkout.CreateDepositRequest{
		ExternalUserID:    "example-user",
		ClientReferenceID: "example-deposit-order",
		SourceAsset:       "USDT",
		TargetAsset:       "USDC",
		Amount:            money.Decimal("100.00"),
	}, client.WithIdempotencyKey("example-deposit-order"))
	if err != nil {
		panic(err)
	}
	fmt.Println(session.CheckoutURL)
}
