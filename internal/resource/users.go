package resource

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/internal/transport"
	"github.com/xautoop/dextri-pay-go/users"
)

type usersService struct{ transport *transport.Client }

// NewUsers builds the user resource implementation used by client.Client.
func NewUsers(client *transport.Client) *usersService {
	return &usersService{transport: client}
}

func (service *usersService) CreateBindingSession(ctx context.Context, request users.CreateBindingSessionRequest, idempotencyKey string) (*users.BindingSession, *api.Response, error) {
	request.ExternalUserID = strings.TrimSpace(request.ExternalUserID)
	if request.ExternalUserID == "" {
		return nil, nil, &api.ValidationError{Field: "external_user_id", Message: "is required"}
	}
	if request.WalletFamily != users.WalletFamilyEVM && request.WalletFamily != users.WalletFamilyCosmos {
		return nil, nil, &api.ValidationError{Field: "wallet_family", Message: "must be evm or cosmos"}
	}
	var output users.BindingSession
	response, err := service.transport.Do(ctx, transport.Request{Method: http.MethodPost, Path: "/v1/user-binding-sessions", Body: request, IdempotencyKey: idempotencyKey}, &output)
	return &output, response, err
}

func (service *usersService) GetBalances(ctx context.Context, userID string) (*users.Balances, *api.Response, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, nil, &api.ValidationError{Field: "external_user_id", Message: "is required"}
	}
	var output users.Balances
	response, err := service.transport.Do(ctx, transport.Request{Method: http.MethodGet, Path: "/v1/users/" + url.PathEscape(userID) + "/balances"}, &output)
	return &output, response, err
}
