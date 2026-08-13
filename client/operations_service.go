package client

import (
	"context"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/operation"
)

type operationsBackend interface {
	Get(context.Context, string) (*operation.Operation, *api.Response, error)
	List(context.Context, operation.ListParams) (*operation.List, *api.Response, error)
}

// OperationsService reads durable App-visible business operations.
type OperationsService struct {
	backend operationsBackend
}

func newOperationsService(backend operationsBackend) *OperationsService {
	return &OperationsService{backend: backend}
}

// Get returns one operation by identifier.
func (service *OperationsService) Get(ctx context.Context, id string) (*operation.Operation, *api.Response, error) {
	return service.backend.Get(ctx, id)
}

// List returns a filtered cursor-paginated operation page.
func (service *OperationsService) List(ctx context.Context, params operation.ListParams) (*operation.List, *api.Response, error) {
	return service.backend.List(ctx, params)
}
