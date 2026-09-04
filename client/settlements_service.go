package client

import (
	"context"
	"net/http"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/escrow"
	"github.com/xautoop/dextri-pay-go/internal/transport"
)

// SettlementsService atomically consumes Escrow funds into explicit destinations.
type SettlementsService struct{ executor executor }

func newSettlementsService(executor executor) *SettlementsService {
	return &SettlementsService{executor: executor}
}

// Create validates atomic conservation and creates one idempotent Settlement.
func (service *SettlementsService) Create(ctx context.Context, request escrow.SettlementRequest, supplied ...RequestOption) (*escrow.Settlement, *api.Response, error) {
	key, err := resolveIdempotencyKey(supplied...)
	if err != nil {
		return nil, nil, err
	}
	if err := request.Validate(); err != nil {
		return nil, nil, err
	}
	var output escrow.Settlement
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodPost, Path: "/v1/escrow-settlements", Body: request, IdempotencyKey: key}, &output)
	return &output, response, err
}
