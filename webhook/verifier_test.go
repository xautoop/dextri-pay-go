package webhook

import (
	"net/http"
	"testing"
	"time"
)

func TestVerifyAndDecode(t *testing.T) {
	now := time.UnixMilli(1700000000000)
	cases := []struct {
		name string
		body string
	}{
		{name: "current contract", body: `{"id":"evt_1","type":"operation.succeeded","created_at":"2023-11-14T22:13:20Z","data":{"operation":{"operation_id":"pop_1","external_user_id":"u","type":"deposit","status":"succeeded","source_asset":"USDT","target_asset":"USDC","input_amount":"10.5","output_amount":"10.4","created_at":"2023-11-14T22:13:20Z","updated_at":"2023-11-14T22:13:20Z"}}}`},
		{name: "legacy outbox contract", body: `{"id":"evt_1","type":"operation.succeeded","created_at":"2023-11-14T22:13:20Z","data":{"operation":{"operation_id":"pop_1","external_user_id":"u","type":"deposit","status":"succeeded","source_asset":"USDT","destination_asset":"USDC","amount":"10.5","output_amount":"10.4","created_at":"2023-11-14T22:13:20Z","updated_at":"2023-11-14T22:13:20Z"}}}`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			body := []byte(test.body)
			headers := http.Header{}
			headers.Set(HeaderDeliveryID, "whd_1")
			headers.Set(HeaderTimestamp, "1700000000000")
			headers.Set(HeaderSignature, sign("secret", "whd_1", "1700000000000", body))
			delivery, err := (Verifier{Secret: "secret", Now: func() time.Time { return now }}).Verify(headers, body)
			if err != nil {
				t.Fatal(err)
			}
			operation := delivery.Event.Data.Operation
			if operation.OperationID != "pop_1" || operation.TargetAsset != "USDC" || operation.InputAmount != "10.5" {
				t.Fatalf("operation=%#v", operation)
			}
		})
	}

	body := []byte(cases[0].body)
	headers := http.Header{}
	headers.Set(HeaderDeliveryID, "whd_1")
	headers.Set(HeaderTimestamp, "1700000000000")
	headers.Set(HeaderSignature, sign("secret", "whd_1", "1700000000000", body))
	if _, err := (Verifier{Secret: "secret", Now: func() time.Time { return now }}).Verify(headers, append(body, ' ')); err == nil {
		t.Fatal("tamper accepted")
	}
}
