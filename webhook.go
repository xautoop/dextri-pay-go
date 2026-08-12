package dextripay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const DefaultWebhookTolerance = 5 * time.Minute

func VerifyWebhook(secret string, headers http.Header, body []byte) error {
	return VerifyWebhookAt(secret, headers, body, time.Now(), DefaultWebhookTolerance)
}

func VerifyWebhookAt(secret string, headers http.Header, body []byte, now time.Time, tolerance time.Duration) error {
	if secret == "" {
		return errors.New("webhook secret is required")
	}
	id := strings.TrimSpace(headers.Get(HeaderWebhookID))
	timestampRaw := strings.TrimSpace(headers.Get(HeaderWebhookTS))
	provided := strings.ToLower(strings.TrimSpace(headers.Get(HeaderWebhookSign)))
	if id == "" || timestampRaw == "" || provided == "" {
		return errors.New("webhook signature headers are incomplete")
	}
	timestampMS, err := strconv.ParseInt(timestampRaw, 10, 64)
	if err != nil {
		return errors.New("webhook timestamp is invalid")
	}
	signedAt := time.UnixMilli(timestampMS)
	delta := now.Sub(signedAt)
	if delta < 0 {
		delta = -delta
	}
	if tolerance <= 0 || delta > tolerance {
		return errors.New("webhook timestamp is outside the allowed window")
	}
	expected := SignWebhook(secret, id, timestampRaw, body)
	if !hmac.Equal([]byte(expected), []byte(provided)) {
		return errors.New("webhook signature verification failed")
	}
	return nil
}

func SignWebhook(secret, webhookID, timestampMS string, body []byte) string {
	payload := strings.Join([]string{webhookSignScheme, strings.TrimSpace(webhookID), strings.TrimSpace(timestampMS), ContentSHA256(body)}, "\n")
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
