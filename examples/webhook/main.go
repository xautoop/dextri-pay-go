package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/xautoop/dextri-pay-go/webhook"
)

func main() {
	http.HandleFunc("/webhooks/dextri-pay", func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(http.MaxBytesReader(writer, request.Body, 4<<20))
		if err != nil {
			http.Error(writer, "invalid body", http.StatusBadRequest)
			return
		}
		delivery, err := webhook.Verify(os.Getenv("DEXTRI_PAY_WEBHOOK_SECRET"), request.Header, body)
		if err != nil {
			http.Error(writer, "invalid signature", http.StatusUnauthorized)
			return
		}
		fmt.Println(delivery.Event.ID, delivery.Event.Type)
		writer.WriteHeader(http.StatusNoContent)
	})
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
