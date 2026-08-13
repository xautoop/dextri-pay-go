package transport

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/xautoop/dextri-pay-go/internal/auth"
)

func (client *Client) send(ctx context.Context, request Request, body []byte) (*http.Response, error) {
	requestURL := *client.baseURL
	escapedPath := strings.TrimRight(client.baseURL.EscapedPath(), "/") + "/" + strings.TrimLeft(request.Path, "/")
	decodedPath, err := url.PathUnescape(escapedPath)
	if err != nil {
		return nil, fmt.Errorf("decode request path: %w", err)
	}
	requestURL.Path, requestURL.RawPath, requestURL.RawQuery = decodedPath, escapedPath, auth.CanonicalQuery(request.Query)
	timestamp := strconv.FormatInt(client.now().UnixMilli(), 10)
	nonce, err := client.nonce()
	if err != nil {
		return nil, fmt.Errorf("create nonce: %w", err)
	}
	contentHash := auth.ContentSHA256(body)
	signature := auth.Sign(client.credentials.Secret, auth.SignInput{
		AppID:             client.credentials.AppID,
		KeyID:             client.credentials.KeyID,
		TimestampMS:       timestamp,
		Nonce:             nonce,
		Method:            request.Method,
		CanonicalResource: auth.CanonicalPathAndQuery(requestURL.EscapedPath(), request.Query),
		ContentSHA256:     contentHash,
	})
	httpRequest, err := http.NewRequestWithContext(ctx, request.Method, requestURL.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set("Accept", "application/json")
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("User-Agent", client.userAgent)
	httpRequest.Header.Set(auth.HeaderAppID, client.credentials.AppID)
	httpRequest.Header.Set(auth.HeaderKeyID, client.credentials.KeyID)
	httpRequest.Header.Set(auth.HeaderTimestamp, timestamp)
	httpRequest.Header.Set(auth.HeaderNonce, nonce)
	httpRequest.Header.Set(auth.HeaderContentSHA, contentHash)
	httpRequest.Header.Set(auth.HeaderSignature, signature)
	if request.IdempotencyKey != "" {
		httpRequest.Header.Set(auth.HeaderIdempotency, request.IdempotencyKey)
	}
	return client.httpClient.Do(httpRequest)
}

func validateBaseURL(raw string, allowInsecure bool) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("base URL must be an absolute HTTPS URL")
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, errors.New("base URL must not contain user info, query or fragment")
	}
	if parsed.Scheme == "https" {
		return parsed, nil
	}
	if parsed.Scheme != "http" || !allowInsecure || !loopbackHost(parsed.Hostname()) {
		return nil, errors.New("plain HTTP is allowed only for an explicitly enabled loopback base URL")
	}
	return parsed, nil
}

func loopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func randomToken() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}
