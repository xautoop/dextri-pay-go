// Package transport owns HTTP execution, authentication and response decoding.
package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/internal/auth"
	"github.com/xautoop/dextri-pay-go/internal/retry"
)

const maxResponseBytes = 4 << 20

type Config struct {
	BaseURL           string
	Credentials       auth.Credentials
	HTTPClient        *http.Client
	RetryPolicy       retry.Policy
	AllowInsecureHTTP bool
	UserAgent         string
	Observer          func(api.Response)
	Now               func() time.Time
	Nonce             func() (string, error)
}

type Request struct {
	Method, Path   string
	Query          url.Values
	Body           any
	IdempotencyKey string
}

type Client struct {
	baseURL     *url.URL
	credentials auth.Credentials
	httpClient  *http.Client
	retry       retry.Policy
	userAgent   string
	observer    func(api.Response)
	now         func() time.Time
	nonce       func() (string, error)
}

func New(config Config) (*Client, error) {
	baseURL, err := validateBaseURL(config.BaseURL, config.AllowInsecureHTTP)
	if err != nil {
		return nil, err
	}
	config.Credentials.AppID = strings.TrimSpace(config.Credentials.AppID)
	config.Credentials.KeyID = strings.TrimSpace(config.Credentials.KeyID)
	if config.Credentials.AppID == "" || config.Credentials.KeyID == "" || config.Credentials.Secret == "" {
		return nil, errors.New("app ID, key ID and secret are required")
	}
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}
	// Signed requests are bound to the original canonical path. Following a
	// redirect would either send an invalid signature to another path or leak
	// Dextri authentication headers to another origin. Clone the caller's
	// client so the SDK can enforce this boundary without mutating shared state.
	httpClient := *config.HTTPClient
	httpClient.CheckRedirect = rejectRedirect
	if config.Now == nil {
		config.Now = time.Now
	}
	if config.Nonce == nil {
		config.Nonce = randomToken
	}
	if strings.TrimSpace(config.UserAgent) == "" {
		config.UserAgent = "dextri-pay-go"
	}
	return &Client{
		baseURL:     baseURL,
		credentials: config.Credentials,
		httpClient:  &httpClient,
		retry:       config.RetryPolicy.Normalize(),
		userAgent:   config.UserAgent,
		observer:    config.Observer,
		now:         config.Now,
		nonce:       config.Nonce,
	}, nil
}

func rejectRedirect(_ *http.Request, _ []*http.Request) error {
	return http.ErrUseLastResponse
}

func (client *Client) Do(ctx context.Context, request Request, output any) (*api.Response, error) {
	body := []byte{}
	var err error
	if request.Body != nil {
		body, err = json.Marshal(request.Body)
		if err != nil {
			return nil, fmt.Errorf("encode request: %w", err)
		}
	}
	request.Method = strings.ToUpper(strings.TrimSpace(request.Method))
	request.IdempotencyKey = strings.TrimSpace(request.IdempotencyKey)
	attempts := 1
	if safeToRetry(request.Method, request.IdempotencyKey) {
		attempts += client.retry.MaxRetries
	}
	var metadata *api.Response
	for attempt := 1; attempt <= attempts; attempt++ {
		response, sendErr := client.send(ctx, request, body)
		if sendErr != nil {
			metadata = &api.Response{IdempotencyKey: request.IdempotencyKey, Attempts: attempt}
			if attempt < attempts && retryableTransportError(sendErr) {
				if err := retry.Wait(ctx, client.retry.Delay(attempt, "", client.now())); err != nil {
					return metadata, err
				}
				continue
			}
			return metadata, &api.RequestError{IdempotencyKey: request.IdempotencyKey, Attempts: attempt, Err: sendErr}
		}
		responseBody, readErr := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
		_ = response.Body.Close()
		metadata = &api.Response{
			StatusCode:     response.StatusCode,
			RequestID:      strings.TrimSpace(response.Header.Get(auth.HeaderRequestID)),
			IdempotencyKey: request.IdempotencyKey,
			Attempts:       attempt,
			Headers:        response.Header.Clone(),
		}
		if client.observer != nil {
			client.observer(*metadata)
		}
		if readErr != nil {
			return metadata, &api.RequestError{RequestID: metadata.RequestID, IdempotencyKey: request.IdempotencyKey, Attempts: attempt, Err: readErr}
		}
		if len(responseBody) > maxResponseBytes {
			return metadata, &api.RequestError{RequestID: metadata.RequestID, IdempotencyKey: request.IdempotencyKey, Attempts: attempt, Err: errors.New("response exceeds 4 MiB")}
		}
		if retryableStatus(response.StatusCode) && attempt < attempts {
			if err := retry.Wait(ctx, client.retry.Delay(attempt, response.Header.Get("Retry-After"), client.now())); err != nil {
				return metadata, err
			}
			continue
		}
		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			return metadata, decodeAPIError(response.StatusCode, metadata.RequestID, request.IdempotencyKey, responseBody)
		}
		if output == nil || len(bytes.TrimSpace(responseBody)) == 0 {
			return metadata, nil
		}
		if err := decodeSuccess(responseBody, output); err != nil {
			return metadata, &api.RequestError{RequestID: metadata.RequestID, IdempotencyKey: request.IdempotencyKey, Attempts: attempt, Err: err}
		}
		return metadata, nil
	}
	return metadata, errors.New("dextri pay request attempts exhausted")
}
