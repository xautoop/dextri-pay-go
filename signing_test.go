package dextripay

import (
	"net/url"
	"testing"
)

func TestRequestSignatureGolden(t *testing.T) {
	body := []byte(`{"amount":"100.00","asset":"USDC"}`)
	contentHash := ContentSHA256(body)
	if contentHash != "785e38d3783454289434e68fafcbd6af03559215eb595622d3e8a13c2d922d0e" {
		t.Fatalf("content hash = %s", contentHash)
	}
	resource := CanonicalPathAndQuery("/v1/checkout-sessions", url.Values{"b": {"3", "2"}, "a": {"1"}})
	if resource != "/v1/checkout-sessions?a=1&b=2&b=3" {
		t.Fatalf("canonical resource = %s", resource)
	}
	signature := Sign("secret_test", SignInput{
		AppID: "app_test", KeyID: "key_001", TimestampMS: "1700000000123", Nonce: "nonce_abc",
		Method: "post", CanonicalResource: resource, ContentSHA256: contentHash,
	})
	if signature != "85ed044ad8c2b3bb618267b28827b11f60285e47ac8a2aede9c8d31c87355544" {
		t.Fatalf("signature = %s", signature)
	}
}
