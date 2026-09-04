package client

import (
	"context"
	"net/http"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/escrow"
	"github.com/xautoop/dextri-pay-go/internal/transport"
)

// EscrowsService atomically commits compatible Holds.
type EscrowsService struct{ executor executor }

func newEscrowsService(executor executor) *EscrowsService { return &EscrowsService{executor: executor} }

// Commit atomically moves all supplied Holds into committed Escrow state.
func (service *EscrowsService) Commit(ctx context.Context, request escrow.CommitRequest, supplied ...RequestOption) (*escrow.Escrow, *api.Response, error) {
	key, err := resolveIdempotencyKey(supplied...)
	if err != nil {
		return nil, nil, err
	}
	if err := request.Validate(); err != nil {
		return nil, nil, err
	}
	var output escrow.Escrow
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodPost, Path: "/v1/escrows/commit", Body: request, IdempotencyKey: key}, &output)
	return &output, response, err
}
