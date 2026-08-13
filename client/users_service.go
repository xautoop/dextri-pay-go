package client

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/internal/transport"
	"github.com/xautoop/dextri-pay-go/users"
)

// UsersService creates wallet-binding sessions and reads live user balances.
type UsersService struct {
	executor executor
}

func newUsersService(executor executor) *UsersService {
	return &UsersService{executor: executor}
}

// CreateBindingSession creates a wallet-binding-only Hosted Checkout session.
func (service *UsersService) CreateBindingSession(ctx context.Context, request users.CreateBindingSessionRequest, supplied ...RequestOption) (*users.BindingSession, *api.Response, error) {
	key, err := resolveIdempotencyKey(supplied...)
	if err != nil {
		return nil, nil, err
	}
	request.ExternalUserID = strings.TrimSpace(request.ExternalUserID)
	if err := request.Validate(); err != nil {
		return nil, nil, err
	}
	var output users.BindingSession
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodPost, Path: "/v1/user-binding-sessions", Body: request, IdempotencyKey: key}, &output)
	return &output, response, err
}

// GetBalances returns live Funding or chain-backed balances for one App user.
func (service *UsersService) GetBalances(ctx context.Context, userID string) (*users.Balances, *api.Response, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, nil, &api.ValidationError{Field: "external_user_id", Message: "is required"}
	}
	var output users.Balances
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodGet, Path: "/v1/users/" + url.PathEscape(userID) + "/balances"}, &output)
	return &output, response, err
}
