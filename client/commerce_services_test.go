package client

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/xautoop/dextri-pay-go/checkout"
	"github.com/xautoop/dextri-pay-go/payment"
	"github.com/xautoop/dextri-pay-go/payout"
)

func TestCommerceServiceContracts(t *testing.T) {
	client := testClient(t, func(request *http.Request) (*http.Response, error) {
		_, _ = io.ReadAll(request.Body)
		body := `{"operation_id":"op_1","external_user_id":"pvsz:user:1","type":"payment","status":"succeeded","created_at":"2023-11-14T22:13:20Z","updated_at":"2023-11-14T22:13:20Z"}`
		if request.URL.Path == "/pay/v1/checkout-sessions" {
			body = `{"session_id":"cs_1","operation_id":"op_1","type":"payment","status":"created","external_user_id":"pvsz:user:1","source_asset":"DXS","target_asset":"DXS","amount":"1","expires_at":"2023-11-14T22:18:20Z"}`
		}
		if request.URL.Path == "/pay/v1/payouts" || request.URL.Path == "/pay/v1/payouts/po_1" {
			body = `{"operation_id":"po_1","external_user_id":"pvsz:user:1","type":"payout","status":"succeeded","created_at":"2023-11-14T22:13:20Z","updated_at":"2023-11-14T22:13:20Z"}`
		}
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(stringsReader(body))}, nil
	})
	ctx := context.Background()
	if _, _, err := client.Checkout.CreatePayment(ctx, checkout.CreatePaymentRequest{ExternalUserID: "pvsz:user:1", ClientReferenceID: "order-1", Asset: "DXS", Amount: "1"}, WithIdempotencyKey("order-0001")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := client.Payments.Get(ctx, "op_1"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := client.Payments.Refund(ctx, "op_1", payment.RefundRequest{RefundID: "refund-1", Reason: "delivery_failed"}, WithIdempotencyKey("refund-0001")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := client.Payouts.Create(ctx, payout.CreateRequest{ExternalUserID: "pvsz:user:1", ClientReferenceID: "reward-1", Asset: "DXS", Amount: "1", Reason: "pvp"}, WithIdempotencyKey("reward-0001")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := client.Payouts.Get(ctx, "po_1"); err != nil {
		t.Fatal(err)
	}
}

type stringReadCloser struct{ value string }

func stringsReader(value string) *stringReadCloser { return &stringReadCloser{value: value} }
func (reader *stringReadCloser) Read(target []byte) (int, error) {
	if reader.value == "" {
		return 0, io.EOF
	}
	n := copy(target, reader.value)
	reader.value = reader.value[n:]
	return n, nil
}

func (reader *stringReadCloser) Close() error { return nil }
