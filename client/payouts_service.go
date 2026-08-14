package client

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/internal/transport"
	"github.com/xautoop/dextri-pay-go/payout"
)

type PayoutsService struct{ executor executor }

func newPayoutsService(executor executor) *PayoutsService { return &PayoutsService{executor: executor} }

func (service *PayoutsService) Create(ctx context.Context, request payout.CreateRequest, supplied ...RequestOption) (*payout.Payout, *api.Response, error) {
	key, err := resolveIdempotencyKey(supplied...)
	if err != nil {
		return nil, nil, err
	}
	if err := request.Validate(); err != nil {
		return nil, nil, err
	}
	var output payout.Payout
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodPost, Path: "/v1/payouts", Body: request, IdempotencyKey: key}, &output)
	return &output, response, err
}

func (service *PayoutsService) Get(ctx context.Context, payoutID string) (*payout.Payout, *api.Response, error) {
	payoutID = strings.TrimSpace(payoutID)
	if payoutID == "" {
		return nil, nil, &api.ValidationError{Field: "payout_id", Message: "is required"}
	}
	var output payout.Payout
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodGet, Path: "/v1/payouts/" + url.PathEscape(payoutID)}, &output)
	return &output, response, err
}
