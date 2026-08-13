package client

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/checkout"
	"github.com/xautoop/dextri-pay-go/internal/transport"
	"github.com/xautoop/dextri-pay-go/money"
)

// CheckoutService creates and reads durable Hosted Checkout sessions.
type CheckoutService struct {
	executor executor
}

func newCheckoutService(executor executor) *CheckoutService {
	return &CheckoutService{executor: executor}
}

// CreateDeposit creates a user-authorized deposit session.
func (service *CheckoutService) CreateDeposit(ctx context.Context, request checkout.CreateDepositRequest, supplied ...RequestOption) (*checkout.Session, *api.Response, error) {
	key, err := resolveIdempotencyKey(supplied...)
	if err != nil {
		return nil, nil, err
	}
	if err := request.Validate(); err != nil {
		return nil, nil, err
	}
	return service.create(ctx, checkout.TypeDeposit, newCheckoutPayload(request.ExternalUserID, request.ClientReferenceID, request.SourceAsset, request.TargetAsset, request.Amount, request.ReturnURL, request.Metadata), key)
}

// CreateWithdrawal creates a user-authorized withdrawal session.
func (service *CheckoutService) CreateWithdrawal(ctx context.Context, request checkout.CreateWithdrawalRequest, supplied ...RequestOption) (*checkout.Session, *api.Response, error) {
	key, err := resolveIdempotencyKey(supplied...)
	if err != nil {
		return nil, nil, err
	}
	if err := request.Validate(); err != nil {
		return nil, nil, err
	}
	return service.create(ctx, checkout.TypeWithdrawal, newCheckoutPayload(request.ExternalUserID, request.ClientReferenceID, request.SourceAsset, request.TargetAsset, request.Amount, request.ReturnURL, request.Metadata), key)
}

// CreateConversion creates a user-authorized chain conversion session.
func (service *CheckoutService) CreateConversion(ctx context.Context, request checkout.CreateConversionRequest, supplied ...RequestOption) (*checkout.Session, *api.Response, error) {
	key, err := resolveIdempotencyKey(supplied...)
	if err != nil {
		return nil, nil, err
	}
	if err := request.Validate(); err != nil {
		return nil, nil, err
	}
	return service.create(ctx, checkout.TypeConversion, newCheckoutPayload(request.ExternalUserID, request.ClientReferenceID, request.SourceAsset, request.TargetAsset, request.Amount, request.ReturnURL, request.Metadata), key)
}

// CreateDepositAndConvert creates a funding session followed by a fresh conversion authorization.
func (service *CheckoutService) CreateDepositAndConvert(ctx context.Context, request checkout.CreateDepositAndConvertRequest, supplied ...RequestOption) (*checkout.Session, *api.Response, error) {
	key, err := resolveIdempotencyKey(supplied...)
	if err != nil {
		return nil, nil, err
	}
	if err := request.Validate(); err != nil {
		return nil, nil, err
	}
	return service.create(ctx, checkout.TypeDepositAndConvert, newCheckoutPayload(request.ExternalUserID, request.ClientReferenceID, request.SourceAsset, request.TargetAsset, request.Amount, request.ReturnURL, request.Metadata), key)
}

// Get returns the latest durable state of a Checkout session.
func (service *CheckoutService) Get(ctx context.Context, id string) (*checkout.Session, *api.Response, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, nil, &api.ValidationError{Field: "session_id", Message: "is required"}
	}
	var output checkout.Session
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodGet, Path: "/v1/checkout-sessions/" + url.PathEscape(id)}, &output)
	return &output, response, err
}

type checkoutPayload struct {
	Type              checkout.Type `json:"type"`
	ExternalUserID    string        `json:"external_user_id"`
	ClientReferenceID string        `json:"client_reference_id,omitempty"`
	SourceAsset       string        `json:"source_asset"`
	TargetAsset       string        `json:"target_asset"`
	Amount            money.Decimal `json:"amount"`
	ReturnURL         string        `json:"return_url,omitempty"`
	Metadata          api.Metadata  `json:"metadata,omitempty"`
}

func newCheckoutPayload(externalUserID, clientReferenceID, sourceAsset, targetAsset string, amount money.Decimal, returnURL string, metadata api.Metadata) checkoutPayload {
	return checkoutPayload{
		ExternalUserID:    strings.TrimSpace(externalUserID),
		ClientReferenceID: strings.TrimSpace(clientReferenceID),
		SourceAsset:       strings.TrimSpace(sourceAsset),
		TargetAsset:       strings.TrimSpace(targetAsset),
		Amount:            amount,
		ReturnURL:         strings.TrimSpace(returnURL),
		Metadata:          metadata,
	}
}

func (service *CheckoutService) create(ctx context.Context, kind checkout.Type, request checkoutPayload, idempotencyKey string) (*checkout.Session, *api.Response, error) {
	request.Type = kind
	var output checkout.Session
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodPost, Path: "/v1/checkout-sessions", Body: request, IdempotencyKey: idempotencyKey}, &output)
	return &output, response, err
}
