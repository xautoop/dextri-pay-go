package dextripay

import (
	"net/http"
	"testing"
	"time"
)

func TestVerifyWebhook(t *testing.T) {
	now := time.UnixMilli(1_700_000_000_000)
	body := []byte(`{"id":"evt_1","type":"deposit.succeeded"}`)
	headers := http.Header{}
	headers.Set(HeaderWebhookID, "whd_1")
	headers.Set(HeaderWebhookTS, "1700000000000")
	headers.Set(HeaderWebhookSign, SignWebhook("whsec_test", "whd_1", "1700000000000", body))
	if err := VerifyWebhookAt("whsec_test", headers, body, now, 5*time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := VerifyWebhookAt("whsec_test", headers, append(body, ' '), now, 5*time.Minute); err == nil {
		t.Fatal("tampered body was accepted")
	}
	if err := VerifyWebhookAt("whsec_test", headers, body, now.Add(6*time.Minute), 5*time.Minute); err == nil {
		t.Fatal("expired webhook was accepted")
	}
}
