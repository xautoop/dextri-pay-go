package auth

import (
	"net/url"
	"strings"
	"testing"
)

func TestRequestSignatureGolden(t *testing.T) {
	body := []byte(`{"amount":"100.00","asset":"USDC"}`)
	hash := ContentSHA256(body)
	if hash != "785e38d3783454289434e68fafcbd6af03559215eb595622d3e8a13c2d922d0e" {
		t.Fatalf("hash=%s", hash)
	}
	resource := CanonicalPathAndQuery("/pay/v1/checkout-sessions", url.Values{"b": {"3", "2"}, "a": {"1"}})
	signature := Sign("secret_test", SignInput{AppID: "app_test", KeyID: "key_001", TimestampMS: "1700000000123", Nonce: "nonce_abc", Method: "post", CanonicalResource: resource, ContentSHA256: hash})
	if signature != "716b85e3250d49243f42ee6ff25fb5132527acb0164297e48358f0902b8e0da8" {
		t.Fatalf("signature=%s", signature)
	}
}

func TestCanonicalQueryEscapesAndDoesNotMutateInput(t *testing.T) {
	query := url.Values{
		"space key": {"a b"},
		"reserved":  {"a/b", "a+b"},
		"empty":     {""},
	}
	if got, want := CanonicalQuery(query), "empty=&reserved=a%2Bb&reserved=a%2Fb&space+key=a+b"; got != want {
		t.Fatalf("CanonicalQuery() = %q, want %q", got, want)
	}
	if got := strings.Join(query["reserved"], ","); got != "a/b,a+b" {
		t.Fatalf("CanonicalQuery mutated input values: %q", got)
	}
}

func FuzzCanonicalPathAndQueryNeverPanics(f *testing.F) {
	f.Add("v1/resource", "key", "value")
	f.Fuzz(func(t *testing.T, path, key, value string) {
		result := CanonicalPathAndQuery(path, url.Values{key: {value}})
		if !strings.HasPrefix(result, "/") {
			t.Fatalf("canonical resource must start with slash: %q", result)
		}
	})
}
