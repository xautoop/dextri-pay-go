package dextripay

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const maxResponseBytes = 4 << 20

type ResponseMeta struct {
	StatusCode int
	RequestID  string
}

type Option func(*Client) error

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) error {
		if httpClient == nil {
			return errors.New("http client is nil")
		}
		c.httpClient = httpClient
		return nil
	}
}

// WithMaxRetries sets retries after the first attempt. Only GET requests and
// requests carrying an idempotency key are retried.
func WithMaxRetries(retries int) Option {
	return func(c *Client) error {
		if retries < 0 || retries > 5 {
			return errors.New("max retries must be between 0 and 5")
		}
		c.maxRetries = retries
		return nil
	}
}

func WithResponseObserver(observer func(ResponseMeta)) Option {
	return func(c *Client) error {
		c.responseObserver = observer
		return nil
	}
}

// WithClock is intended for deterministic tests and controlled runtimes.
func WithClock(clock func() time.Time) Option {
	return func(c *Client) error {
		if clock == nil {
			return errors.New("clock is nil")
		}
		c.now = clock
		return nil
	}
}

// WithNonceSource is intended for deterministic tests. Production callers
// should use the cryptographically random default.
func WithNonceSource(source func() (string, error)) Option {
	return func(c *Client) error {
		if source == nil {
			return errors.New("nonce source is nil")
		}
		c.nonce = source
		return nil
	}
}

type Client struct {
	baseURL          *url.URL
	httpClient       *http.Client
	appID            string
	keyID            string
	secret           string
	maxRetries       int
	retryWait        time.Duration
	now              func() time.Time
	nonce            func() (string, error)
	responseObserver func(ResponseMeta)

	Channels    *ChannelsService
	Checkout    *CheckoutService
	Users       *UsersService
	Conversions *ConversionsService
	Operations  *OperationsService
}

func (c *Client) String() string {
	if c == nil {
		return "dextripay.Client<nil>"
	}
	return fmt.Sprintf("dextripay.Client{base_url:%q, app_id:%q, key_id:%q, secret:[REDACTED]}", c.baseURL, c.appID, c.keyID)
}

func (c *Client) GoString() string { return c.String() }

func NewClient(baseURL, appID, keyID, secret string, options ...Option) (*Client, error) {
	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("base URL must be an absolute HTTP(S) URL")
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return nil, errors.New("base URL must use HTTP or HTTPS")
	}
	if strings.TrimSpace(appID) == "" || strings.TrimSpace(keyID) == "" || secret == "" {
		return nil, errors.New("app ID, key ID and secret are required")
	}
	c := &Client{
		baseURL: parsed, appID: strings.TrimSpace(appID), keyID: strings.TrimSpace(keyID), secret: secret,
		httpClient: &http.Client{Timeout: 30 * time.Second}, maxRetries: 2, retryWait: 100 * time.Millisecond,
		now: time.Now, nonce: randomToken,
	}
	for _, option := range options {
		if option != nil {
			if err := option(c); err != nil {
				return nil, err
			}
		}
	}
	c.Channels = &ChannelsService{client: c}
	c.Checkout = &CheckoutService{client: c}
	c.Users = &UsersService{client: c}
	c.Conversions = &ConversionsService{client: c}
	c.Operations = &OperationsService{client: c}
	return c, nil
}

func NewIdempotencyKey() (string, error) { return prefixedRandomToken("idem_") }

func randomToken() (string, error) { return prefixedRandomToken("") }

func prefixedRandomToken(prefix string) (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(raw), nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, query url.Values, payload, output any, idempotencyKey string) error {
	body := []byte{}
	var err error
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
	}
	if methodRequiresIdempotency(method) && strings.TrimSpace(idempotencyKey) == "" {
		idempotencyKey, err = NewIdempotencyKey()
		if err != nil {
			return fmt.Errorf("create idempotency key: %w", err)
		}
	}
	return c.execute(ctx, method, path, query, body, output, strings.TrimSpace(idempotencyKey))
}

func (c *Client) execute(ctx context.Context, method, path string, query url.Values, body []byte, output any, idempotencyKey string) error {
	attempts := 1
	if method == http.MethodGet || method == http.MethodHead || idempotencyKey != "" {
		attempts += c.maxRetries
	}
	for attempt := 0; attempt < attempts; attempt++ {
		resp, err := c.send(ctx, method, path, query, body, idempotencyKey)
		if err != nil {
			if attempt+1 < attempts && retryableTransportError(err) {
				if waitErr := waitContext(ctx, c.retryWait*time.Duration(attempt+1)); waitErr != nil {
					return waitErr
				}
				continue
			}
			return &RequestError{Err: err}
		}
		responseBody, readErr := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
		_ = resp.Body.Close()
		meta := ResponseMeta{StatusCode: resp.StatusCode, RequestID: strings.TrimSpace(resp.Header.Get(HeaderRequestID))}
		if c.responseObserver != nil {
			c.responseObserver(meta)
		}
		if readErr != nil {
			return &RequestError{RequestID: meta.RequestID, Err: readErr}
		}
		if len(responseBody) > maxResponseBytes {
			return &RequestError{RequestID: meta.RequestID, Err: errors.New("response exceeds 4 MiB")}
		}
		if retryableStatus(resp.StatusCode) && attempt+1 < attempts {
			if waitErr := waitContext(ctx, c.retryWait*time.Duration(attempt+1)); waitErr != nil {
				return waitErr
			}
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return decodeAPIError(resp.StatusCode, meta.RequestID, responseBody)
		}
		if output == nil || len(bytes.TrimSpace(responseBody)) == 0 {
			return nil
		}
		if err := decodeSuccess(responseBody, output); err != nil {
			return &RequestError{RequestID: meta.RequestID, Err: err}
		}
		return nil
	}
	return errors.New("dextri pay request attempts exhausted")
}

func (c *Client) send(ctx context.Context, method, path string, query url.Values, body []byte, idempotencyKey string) (*http.Response, error) {
	requestURL := *c.baseURL
	escapedPath := strings.TrimRight(c.baseURL.EscapedPath(), "/") + "/" + strings.TrimLeft(path, "/")
	decodedPath, err := url.PathUnescape(escapedPath)
	if err != nil {
		return nil, fmt.Errorf("decode request path: %w", err)
	}
	requestURL.Path = decodedPath
	requestURL.RawPath = escapedPath
	requestURL.RawQuery = canonicalQuery(query)
	timestamp := strconv.FormatInt(c.now().UnixMilli(), 10)
	nonce, err := c.nonce()
	if err != nil {
		return nil, fmt.Errorf("create nonce: %w", err)
	}
	contentHash := ContentSHA256(body)
	canonicalResource := CanonicalPathAndQuery(requestURL.EscapedPath(), query)
	signature := Sign(c.secret, SignInput{
		AppID: c.appID, KeyID: c.keyID, TimestampMS: timestamp, Nonce: nonce,
		Method: method, CanonicalResource: canonicalResource, ContentSHA256: contentHash,
	})
	req, err := http.NewRequestWithContext(ctx, method, requestURL.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(HeaderAppID, c.appID)
	req.Header.Set(HeaderKeyID, c.keyID)
	req.Header.Set(HeaderTimestamp, timestamp)
	req.Header.Set(HeaderNonce, nonce)
	req.Header.Set(HeaderContentSHA, contentHash)
	req.Header.Set(HeaderSignature, signature)
	if idempotencyKey != "" {
		req.Header.Set(HeaderIdempotency, idempotencyKey)
	}
	return c.httpClient.Do(req)
}

func decodeSuccess(body []byte, output any) error {
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err == nil && len(envelope.Data) > 0 && string(envelope.Data) != "null" {
		return json.Unmarshal(envelope.Data, output)
	}
	return json.Unmarshal(body, output)
}

func decodeAPIError(statusCode int, requestID string, body []byte) error {
	apiErr := &APIError{StatusCode: statusCode, RequestID: requestID, Message: http.StatusText(statusCode)}
	var envelope struct {
		Code      string         `json:"code"`
		Message   string         `json:"message"`
		Msg       string         `json:"msg"`
		RequestID string         `json:"request_id"`
		Details   map[string]any `json:"details"`
		Error     *struct {
			Code      string         `json:"code"`
			Message   string         `json:"message"`
			RequestID string         `json:"request_id"`
			Details   map[string]any `json:"details"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &envelope) == nil {
		apiErr.Code, apiErr.Message, apiErr.Details = envelope.Code, firstNonEmpty(envelope.Message, envelope.Msg, apiErr.Message), envelope.Details
		apiErr.RequestID = firstNonEmpty(apiErr.RequestID, envelope.RequestID)
		if envelope.Error != nil {
			apiErr.Code, apiErr.Message, apiErr.Details = envelope.Error.Code, firstNonEmpty(envelope.Error.Message, apiErr.Message), envelope.Error.Details
			apiErr.RequestID = firstNonEmpty(apiErr.RequestID, envelope.Error.RequestID)
		}
	}
	return apiErr
}

func methodRequiresIdempotency(method string) bool {
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch
}

func retryableStatus(status int) bool {
	return status == http.StatusRequestTimeout || status == http.StatusTooManyRequests || status == http.StatusInternalServerError ||
		status == http.StatusBadGateway || status == http.StatusServiceUnavailable || status == http.StatusGatewayTimeout
}

func retryableTransportError(err error) bool {
	return !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded)
}

func waitContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
