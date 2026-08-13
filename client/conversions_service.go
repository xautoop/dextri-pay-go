package client

import (
	"context"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/conversion"
)

type conversionsBackend interface {
	ListMarkets(context.Context) (*conversion.MarketList, *api.Response, error)
	GetMarket(context.Context, string) (*conversion.Market, *api.Response, error)
	UpdatePrice(context.Context, string, conversion.UpdatePriceRequest, string) (*conversion.Market, *api.Response, error)
	CreateQuote(context.Context, conversion.CreateQuoteRequest, string) (*conversion.Quote, *api.Response, error)
}

// ConversionsService reads dynamic markets, updates App prices and creates quotes.
type ConversionsService struct {
	backend conversionsBackend
}

func newConversionsService(backend conversionsBackend) *ConversionsService {
	return &ConversionsService{backend: backend}
}

// ListMarkets returns all Admin-configured markets for the authenticated App.
func (service *ConversionsService) ListMarkets(ctx context.Context) (*conversion.MarketList, *api.Response, error) {
	return service.backend.ListMarkets(ctx)
}

// GetMarket returns one chain-registry-backed market snapshot.
func (service *ConversionsService) GetMarket(ctx context.Context, id string) (*conversion.Market, *api.Response, error) {
	return service.backend.GetMarket(ctx, id)
}

// UpdatePrice atomically replaces both App-defined sides of one market price.
func (service *ConversionsService) UpdatePrice(ctx context.Context, id string, request conversion.UpdatePriceRequest, supplied ...RequestOption) (*conversion.Market, *api.Response, error) {
	key, err := resolveIdempotencyKey(supplied...)
	if err != nil {
		return nil, nil, err
	}
	return service.backend.UpdatePrice(ctx, id, request, key)
}

// CreateQuote creates a short-lived user-signable quote.
func (service *ConversionsService) CreateQuote(ctx context.Context, request conversion.CreateQuoteRequest, supplied ...RequestOption) (*conversion.Quote, *api.Response, error) {
	key, err := resolveIdempotencyKey(supplied...)
	if err != nil {
		return nil, nil, err
	}
	return service.backend.CreateQuote(ctx, request, key)
}
