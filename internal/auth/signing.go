// Package auth implements the Dextri Pay request-signing protocol.
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"sort"
	"strings"
)

const (
	SignatureScheme   = "DEXTRI-PAY-HMAC-SHA256"
	HeaderAppID       = "X-DEXTRI-PAY-APP-ID"
	HeaderKeyID       = "X-DEXTRI-PAY-KEY-ID"
	HeaderTimestamp   = "X-DEXTRI-PAY-TS"
	HeaderNonce       = "X-DEXTRI-PAY-NONCE"
	HeaderContentSHA  = "X-DEXTRI-PAY-CONTENT-SHA256"
	HeaderSignature   = "X-DEXTRI-PAY-SIGN"
	HeaderIdempotency = "Idempotency-Key"
	HeaderRequestID   = "X-Request-ID"
)

type Credentials struct {
	AppID  string
	KeyID  string
	Secret string
}

type SignInput struct {
	AppID             string
	KeyID             string
	TimestampMS       string
	Nonce             string
	Method            string
	CanonicalResource string
	ContentSHA256     string
}

func CanonicalPayload(input SignInput) string {
	return strings.Join([]string{
		SignatureScheme,
		strings.TrimSpace(input.AppID),
		strings.TrimSpace(input.KeyID),
		strings.TrimSpace(input.TimestampMS),
		strings.TrimSpace(input.Nonce),
		strings.ToUpper(strings.TrimSpace(input.Method)),
		input.CanonicalResource,
		strings.ToLower(strings.TrimSpace(input.ContentSHA256)),
	}, "\n")
}

func Sign(secret string, input SignInput) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(CanonicalPayload(input)))
	return hex.EncodeToString(mac.Sum(nil))
}

func ContentSHA256(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

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
	return path + "?" + CanonicalQuery(query)
}

func CanonicalQuery(query url.Values) string {
	normalized := make(url.Values, len(query))
	for key, values := range query {
		copied := append([]string(nil), values...)
		sort.Strings(copied)
		normalized[key] = copied
	}
	return normalized.Encode()
}
