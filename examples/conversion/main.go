package main

import (
	"context"
	"fmt"
	"os"

	"github.com/xautoop/dextri-pay-go/client"
	"github.com/xautoop/dextri-pay-go/conversion"
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
		UserAgent: "dextri-pay-conversion-example/1.0.0",
	})
	if err != nil {
		panic(err)
	}
	quote, _, err := pay.Conversions.CreateQuote(context.Background(), conversion.CreateQuoteRequest{
		ExternalUserID: "example-user",
		MarketID:       os.Getenv("DEXTRI_PAY_MARKET_ID"),
		Side:           conversion.SideSellBase,
		InputAmount:    money.Decimal("25.5"),
	}, client.WithIdempotencyKey("example-conversion-order"))
	if err != nil {
		panic(err)
	}
	fmt.Println(quote.OutputAmount)
}
