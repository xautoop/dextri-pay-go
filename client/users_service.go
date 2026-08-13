package client

import (
	"context"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/users"
)

type usersBackend interface {
	CreateBindingSession(context.Context, users.CreateBindingSessionRequest, string) (*users.BindingSession, *api.Response, error)
	GetBalances(context.Context, string) (*users.Balances, *api.Response, error)
}

// UsersService creates wallet-binding sessions and reads live user balances.
type UsersService struct {
	backend usersBackend
}

func newUsersService(backend usersBackend) *UsersService {
	return &UsersService{backend: backend}
}

// CreateBindingSession creates a wallet-binding-only Hosted Checkout session.
func (service *UsersService) CreateBindingSession(ctx context.Context, request users.CreateBindingSessionRequest, supplied ...RequestOption) (*users.BindingSession, *api.Response, error) {
	key, err := resolveIdempotencyKey(supplied...)
	if err != nil {
		return nil, nil, err
	}
	return service.backend.CreateBindingSession(ctx, request, key)
}

// GetBalances returns live Funding or chain-backed balances for one App user.
func (service *UsersService) GetBalances(ctx context.Context, userID string) (*users.Balances, *api.Response, error) {
	return service.backend.GetBalances(ctx, userID)
}
