package client

import (
	"context"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/checkout"
)

type checkoutBackend interface {
	CreateDeposit(context.Context, checkout.CreateDepositRequest, string) (*checkout.Session, *api.Response, error)
	CreateWithdrawal(context.Context, checkout.CreateWithdrawalRequest, string) (*checkout.Session, *api.Response, error)
	CreateConversion(context.Context, checkout.CreateConversionRequest, string) (*checkout.Session, *api.Response, error)
	CreateDepositAndConvert(context.Context, checkout.CreateDepositAndConvertRequest, string) (*checkout.Session, *api.Response, error)
	Get(context.Context, string) (*checkout.Session, *api.Response, error)
}

// CheckoutService creates and reads durable Hosted Checkout sessions.
type CheckoutService struct {
	backend checkoutBackend
}

func newCheckoutService(backend checkoutBackend) *CheckoutService {
	return &CheckoutService{backend: backend}
}

// CreateDeposit creates a user-authorized deposit session.
func (service *CheckoutService) CreateDeposit(ctx context.Context, request checkout.CreateDepositRequest, supplied ...RequestOption) (*checkout.Session, *api.Response, error) {
	key, err := resolveIdempotencyKey(supplied...)
	if err != nil {
		return nil, nil, err
	}
	return service.backend.CreateDeposit(ctx, request, key)
}

// CreateWithdrawal creates a user-authorized withdrawal session.
func (service *CheckoutService) CreateWithdrawal(ctx context.Context, request checkout.CreateWithdrawalRequest, supplied ...RequestOption) (*checkout.Session, *api.Response, error) {
	key, err := resolveIdempotencyKey(supplied...)
	if err != nil {
		return nil, nil, err
	}
	return service.backend.CreateWithdrawal(ctx, request, key)
}

// CreateConversion creates a user-authorized chain conversion session.
func (service *CheckoutService) CreateConversion(ctx context.Context, request checkout.CreateConversionRequest, supplied ...RequestOption) (*checkout.Session, *api.Response, error) {
	key, err := resolveIdempotencyKey(supplied...)
	if err != nil {
		return nil, nil, err
	}
	return service.backend.CreateConversion(ctx, request, key)
}

// CreateDepositAndConvert creates a funding session followed by a fresh conversion authorization.
func (service *CheckoutService) CreateDepositAndConvert(ctx context.Context, request checkout.CreateDepositAndConvertRequest, supplied ...RequestOption) (*checkout.Session, *api.Response, error) {
	key, err := resolveIdempotencyKey(supplied...)
	if err != nil {
		return nil, nil, err
	}
	return service.backend.CreateDepositAndConvert(ctx, request, key)
}

// Get returns the latest durable state of a Checkout session.
func (service *CheckoutService) Get(ctx context.Context, id string) (*checkout.Session, *api.Response, error) {
	return service.backend.Get(ctx, id)
}
