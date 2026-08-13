package client_test

import (
	"fmt"
	"strings"
	"time"

	"github.com/xautoop/dextri-pay-go/client"
)

func ExampleNew() {
	pay, err := client.New(client.Config{
		BaseURL: "https://pay.example",
		Credentials: client.Credentials{
			AppID:  "app_example",
			KeyID:  "key_example",
			Secret: "server-side-secret",
		},
	}, client.WithRetryPolicy(2, 200*time.Millisecond, 5*time.Second))

	fmt.Println(err == nil)
	fmt.Println(pay.Checkout != nil && pay.Operations != nil)
	// Output:
	// true
	// true
}

func ExampleNewIdempotencyKey() {
	key, err := client.NewIdempotencyKey()
	fmt.Println(err == nil)
	fmt.Println(strings.HasPrefix(key, "idem_"))
	// Output:
	// true
	// true
}
