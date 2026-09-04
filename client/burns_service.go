package client

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/burn"
	"github.com/xautoop/dextri-pay-go/internal/transport"
)

// BurnsService creates and reads auditable App-account destruction requests.
type BurnsService struct{ executor executor }

func newBurnsService(executor executor) *BurnsService { return &BurnsService{executor: executor} }

// Create validates and submits one idempotent burn request.
func (service *BurnsService) Create(ctx context.Context, request burn.CreateRequest, supplied ...RequestOption) (*burn.Burn, *api.Response, error) {
	key, err := resolveIdempotencyKey(supplied...)
	if err != nil {
		return nil, nil, err
	}
	if err := request.Validate(); err != nil {
		return nil, nil, err
	}
	var output burn.Burn
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodPost, Path: "/v1/burns", Body: request, IdempotencyKey: key}, &output)
	return &output, response, err
}

// Get reads the latest authoritative burn state.
func (service *BurnsService) Get(ctx context.Context, burnID string) (*burn.Burn, *api.Response, error) {
	burnID = strings.TrimSpace(burnID)
	if burnID == "" {
		return nil, nil, &api.ValidationError{Field: "burn_id", Message: "is required"}
	}
	var output burn.Burn
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodGet, Path: "/v1/burns/" + url.PathEscape(burnID)}, &output)
	return &output, response, err
}
