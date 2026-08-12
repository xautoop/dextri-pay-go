package dextripay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"sort"
	"strings"
)

const (
	SignatureScheme = "DEXTRI-PAY-HMAC-SHA256"

	HeaderAppID       = "X-DEXTRI-PAY-APP-ID"
	HeaderKeyID       = "X-DEXTRI-PAY-KEY-ID"
	HeaderTimestamp   = "X-DEXTRI-PAY-TS"
	HeaderNonce       = "X-DEXTRI-PAY-NONCE"
	HeaderContentSHA  = "X-DEXTRI-PAY-CONTENT-SHA256"
	HeaderSignature   = "X-DEXTRI-PAY-SIGN"
	HeaderIdempotency = "Idempotency-Key"
	HeaderRequestID   = "X-Request-ID"
	HeaderWebhookID   = "X-DEXTRI-PAY-WEBHOOK-ID"
	HeaderWebhookTS   = "X-DEXTRI-PAY-WEBHOOK-TS"
	HeaderWebhookSign = "X-DEXTRI-PAY-WEBHOOK-SIGN"
	webhookSignScheme = "DEXTRI-PAY-WEBHOOK-HMAC-SHA256"
)

// SignInput contains the exact request fields authenticated by App HMAC.
type SignInput struct {
	AppID             string
	KeyID             string
	TimestampMS       string
	Nonce             string
	Method            string
	CanonicalResource string
	ContentSHA256     string
}

// CanonicalPayload returns the newline-delimited HMAC payload.
func CanonicalPayload(in SignInput) string {
	return strings.Join([]string{
		SignatureScheme,
		strings.TrimSpace(in.AppID),
		strings.TrimSpace(in.KeyID),
		strings.TrimSpace(in.TimestampMS),
		strings.TrimSpace(in.Nonce),
		strings.ToUpper(strings.TrimSpace(in.Method)),
		in.CanonicalResource,
		strings.ToLower(strings.TrimSpace(in.ContentSHA256)),
	}, "\n")
}

// Sign returns the lowercase hexadecimal HMAC-SHA256 signature.
func Sign(secret string, in SignInput) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(CanonicalPayload(in)))
	return hex.EncodeToString(mac.Sum(nil))
}

// ContentSHA256 returns the lowercase hexadecimal SHA-256 of body.
func ContentSHA256(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

// CanonicalPathAndQuery normalizes a path and sorts query keys and values.
func CanonicalPathAndQuery(path string, query url.Values) string {
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if len(query) == 0 {
		return path
	}
	return path + "?" + canonicalQuery(query)
}

func canonicalQuery(query url.Values) string {
	normalized := make(url.Values, len(query))
	for key, values := range query {
		copied := append([]string(nil), values...)
		sort.Strings(copied)
		normalized[key] = copied
	}
	return normalized.Encode()
}
