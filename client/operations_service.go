package client

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/internal/transport"
	"github.com/xautoop/dextri-pay-go/operation"
)

// OperationsService reads durable App-visible business operations.
type OperationsService struct {
	executor executor
}

func newOperationsService(executor executor) *OperationsService {
	return &OperationsService{executor: executor}
}

// Get returns one operation by identifier.
func (service *OperationsService) Get(ctx context.Context, id string) (*operation.Operation, *api.Response, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, nil, &api.ValidationError{Field: "operation_id", Message: "is required"}
	}
	var output operation.Operation
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodGet, Path: "/v1/operations/" + url.PathEscape(id)}, &output)
	return &output, response, err
}

// List returns a filtered cursor-paginated operation page.
func (service *OperationsService) List(ctx context.Context, params operation.ListParams) (*operation.List, *api.Response, error) {
	if err := params.Validate(); err != nil {
		return nil, nil, err
	}
	query := url.Values{}
	if value := strings.TrimSpace(params.ExternalUserID); value != "" {
		query.Set("external_user_id", value)
	}
	if params.Type != "" {
		query.Set("type", string(params.Type))
	}
	if params.Status != "" {
		query.Set("status", string(params.Status))
	}
	if value := strings.TrimSpace(params.Cursor); value != "" {
		query.Set("cursor", value)
	}
	if params.Limit > 0 {
		query.Set("limit", strconv.Itoa(params.Limit))
	}
	var output operation.List
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodGet, Path: "/v1/operations", Query: query}, &output)
	return &output, response, err
}
