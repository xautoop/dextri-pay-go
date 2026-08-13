package client

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/conversion"
	"github.com/xautoop/dextri-pay-go/internal/transport"
)

// ConversionsService reads dynamic markets, updates App prices and creates quotes.
type ConversionsService struct {
	executor executor
}

func newConversionsService(executor executor) *ConversionsService {
	return &ConversionsService{executor: executor}
}

// ListMarkets returns all Admin-configured markets for the authenticated App.
func (service *ConversionsService) ListMarkets(ctx context.Context) (*conversion.MarketList, *api.Response, error) {
	var output conversion.MarketList
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodGet, Path: "/v1/conversion-markets"}, &output)
	return &output, response, err
}

// GetMarket returns one chain-registry-backed market snapshot.
func (service *ConversionsService) GetMarket(ctx context.Context, id string) (*conversion.Market, *api.Response, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, nil, &api.ValidationError{Field: "market_id", Message: "is required"}
	}
	var output conversion.Market
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodGet, Path: "/v1/conversion-markets/" + url.PathEscape(id)}, &output)
	return &output, response, err
}

// UpdatePrice atomically replaces both App-defined sides of one market price.
func (service *ConversionsService) UpdatePrice(ctx context.Context, id string, request conversion.UpdatePriceRequest, supplied ...RequestOption) (*conversion.Market, *api.Response, error) {
	key, err := resolveIdempotencyKey(supplied...)
	if err != nil {
		return nil, nil, err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, nil, &api.ValidationError{Field: "market_id", Message: "is required"}
	}
	if err := request.Validate(); err != nil {
		return nil, nil, &api.ValidationError{Field: "conversion_price", Message: err.Error()}
	}
	var output conversion.Market
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodPut, Path: "/v1/conversion-markets/" + url.PathEscape(id) + "/price", Body: request, IdempotencyKey: key}, &output)
	return &output, response, err
}

// CreateQuote creates a short-lived user-signable quote.
func (service *ConversionsService) CreateQuote(ctx context.Context, request conversion.CreateQuoteRequest, supplied ...RequestOption) (*conversion.Quote, *api.Response, error) {
	key, err := resolveIdempotencyKey(supplied...)
	if err != nil {
		return nil, nil, err
	}
	request.ExternalUserID = strings.TrimSpace(request.ExternalUserID)
	request.Owner = strings.TrimSpace(request.Owner)
	request.MarketID = strings.TrimSpace(request.MarketID)
	if err := request.Validate(); err != nil {
		return nil, nil, err
	}
	var output conversion.Quote
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodPost, Path: "/v1/conversion-quotes", Body: request, IdempotencyKey: key}, &output)
	return &output, response, err
}
