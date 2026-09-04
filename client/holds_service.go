package client

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/hold"
	"github.com/xautoop/dextri-pay-go/internal/transport"
)

// HoldsService creates, reads and releases generic balance reservations.
type HoldsService struct{ executor executor }

func newHoldsService(executor executor) *HoldsService { return &HoldsService{executor: executor} }

// Create reserves available user balance under an explicit idempotency key.
func (service *HoldsService) Create(ctx context.Context, request hold.CreateRequest, supplied ...RequestOption) (*hold.Hold, *api.Response, error) {
	key, err := resolveIdempotencyKey(supplied...)
	if err != nil {
		return nil, nil, err
	}
	if err := request.Validate(); err != nil {
		return nil, nil, err
	}
	var output hold.Hold
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodPost, Path: "/v1/holds", Body: request, IdempotencyKey: key}, &output)
	return &output, response, err
}

// Get reads the latest authoritative Hold state.
func (service *HoldsService) Get(ctx context.Context, holdID string) (*hold.Hold, *api.Response, error) {
	holdID = strings.TrimSpace(holdID)
	if holdID == "" {
		return nil, nil, &api.ValidationError{Field: "hold_id", Message: "is required"}
	}
	var output hold.Hold
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodGet, Path: "/v1/holds/" + url.PathEscape(holdID)}, &output)
	return &output, response, err
}

// Release releases an unconsumed Hold under an explicit idempotency key.
func (service *HoldsService) Release(ctx context.Context, holdID string, request hold.ReleaseRequest, supplied ...RequestOption) (*hold.Hold, *api.Response, error) {
	holdID = strings.TrimSpace(holdID)
	if holdID == "" {
		return nil, nil, &api.ValidationError{Field: "hold_id", Message: "is required"}
	}
	key, err := resolveIdempotencyKey(supplied...)
	if err != nil {
		return nil, nil, err
	}
	if err := request.Validate(); err != nil {
		return nil, nil, err
	}
	var output hold.Hold
	path := "/v1/holds/" + url.PathEscape(holdID) + "/release"
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodPost, Path: path, Body: request, IdempotencyKey: key}, &output)
	return &output, response, err
}
