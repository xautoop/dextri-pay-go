package resource

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

type operationsService struct{ transport *transport.Client }

// NewOperations builds the operation resource implementation used by client.Client.
func NewOperations(client *transport.Client) *operationsService {
	return &operationsService{transport: client}
}

func (service *operationsService) Get(ctx context.Context, id string) (*operation.Operation, *api.Response, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, nil, &api.ValidationError{Field: "operation_id", Message: "is required"}
	}
	var output operation.Operation
	response, err := service.transport.Do(ctx, transport.Request{Method: http.MethodGet, Path: "/v1/operations/" + url.PathEscape(id)}, &output)
	return &output, response, err
}

func (service *operationsService) List(ctx context.Context, params operation.ListParams) (*operation.List, *api.Response, error) {
	if params.Limit < 0 {
		return nil, nil, &api.ValidationError{Field: "limit", Message: "must not be negative"}
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
	response, err := service.transport.Do(ctx, transport.Request{Method: http.MethodGet, Path: "/v1/operations", Query: query}, &output)
	return &output, response, err
}
