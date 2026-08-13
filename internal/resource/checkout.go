package resource

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

type checkoutService struct{ transport *transport.Client }

// NewCheckout builds the Hosted Checkout implementation used by client.Client.
func NewCheckout(client *transport.Client) *checkoutService {
	return &checkoutService{transport: client}
}

func (service *checkoutService) CreateDeposit(ctx context.Context, request checkout.CreateDepositRequest, idempotencyKey string) (*checkout.Session, *api.Response, error) {
	return service.create(ctx, checkout.TypeDeposit, checkoutPayload{ExternalUserID: request.ExternalUserID, ClientReferenceID: request.ClientReferenceID, SourceAsset: request.SourceAsset, TargetAsset: request.TargetAsset, Amount: request.Amount, ReturnURL: request.ReturnURL, Metadata: request.Metadata}, idempotencyKey)
}

func (service *checkoutService) CreateWithdrawal(ctx context.Context, request checkout.CreateWithdrawalRequest, idempotencyKey string) (*checkout.Session, *api.Response, error) {
	return service.create(ctx, checkout.TypeWithdrawal, checkoutPayload{ExternalUserID: request.ExternalUserID, ClientReferenceID: request.ClientReferenceID, SourceAsset: request.SourceAsset, TargetAsset: request.TargetAsset, Amount: request.Amount, ReturnURL: request.ReturnURL, Metadata: request.Metadata}, idempotencyKey)
}

func (service *checkoutService) CreateConversion(ctx context.Context, request checkout.CreateConversionRequest, idempotencyKey string) (*checkout.Session, *api.Response, error) {
	return service.create(ctx, checkout.TypeConversion, checkoutPayload{ExternalUserID: request.ExternalUserID, ClientReferenceID: request.ClientReferenceID, SourceAsset: request.SourceAsset, TargetAsset: request.TargetAsset, Amount: request.Amount, ReturnURL: request.ReturnURL, Metadata: request.Metadata}, idempotencyKey)
}

func (service *checkoutService) CreateDepositAndConvert(ctx context.Context, request checkout.CreateDepositAndConvertRequest, idempotencyKey string) (*checkout.Session, *api.Response, error) {
	return service.create(ctx, checkout.TypeDepositAndConvert, checkoutPayload{ExternalUserID: request.ExternalUserID, ClientReferenceID: request.ClientReferenceID, SourceAsset: request.SourceAsset, TargetAsset: request.TargetAsset, Amount: request.Amount, ReturnURL: request.ReturnURL, Metadata: request.Metadata}, idempotencyKey)
}

func (service *checkoutService) Get(ctx context.Context, id string) (*checkout.Session, *api.Response, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, nil, &api.ValidationError{Field: "session_id", Message: "is required"}
	}
	var output checkout.Session
	response, err := service.transport.Do(ctx, transport.Request{Method: http.MethodGet, Path: "/v1/checkout-sessions/" + url.PathEscape(id)}, &output)
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

func (service *checkoutService) create(ctx context.Context, kind checkout.Type, request checkoutPayload, idempotencyKey string) (*checkout.Session, *api.Response, error) {
	request.Type = kind
	request.ExternalUserID = strings.TrimSpace(request.ExternalUserID)
	request.ClientReferenceID = strings.TrimSpace(request.ClientReferenceID)
	request.SourceAsset = strings.TrimSpace(request.SourceAsset)
	request.TargetAsset = strings.TrimSpace(request.TargetAsset)
	request.ReturnURL = strings.TrimSpace(request.ReturnURL)
	if request.ExternalUserID == "" {
		return nil, nil, &api.ValidationError{Field: "external_user_id", Message: "is required"}
	}
	if request.SourceAsset == "" || request.TargetAsset == "" {
		return nil, nil, &api.ValidationError{Field: "asset", Message: "source_asset and target_asset are required"}
	}
	if err := request.Amount.ValidatePositive(); err != nil {
		return nil, nil, &api.ValidationError{Field: "amount", Message: err.Error()}
	}
	var output checkout.Session
	response, err := service.transport.Do(ctx, transport.Request{Method: http.MethodPost, Path: "/v1/checkout-sessions", Body: request, IdempotencyKey: idempotencyKey}, &output)
	return &output, response, err
}
