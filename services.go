package dextripay

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

type ChannelsService struct{ client *Client }

func (s *ChannelsService) List(ctx context.Context, params ListChannelsParams) ([]Channel, error) {
	query := url.Values{}
	if params.Flow != "" {
		query.Set("flow", params.Flow)
	}
	if params.SourceAsset != "" {
		query.Set("source_asset", params.SourceAsset)
	}
	var out []Channel
	err := s.client.doJSON(ctx, http.MethodGet, "/v1/channels", query, nil, &out, "")
	return out, err
}

type CheckoutService struct{ client *Client }

func (s *CheckoutService) CreateDeposit(ctx context.Context, request CreateCheckoutRequest) (*CheckoutSession, error) {
	return s.create(ctx, "deposit", request)
}

func (s *CheckoutService) CreateWithdrawal(ctx context.Context, request CreateCheckoutRequest) (*CheckoutSession, error) {
	return s.create(ctx, "withdrawal", request)
}

func (s *CheckoutService) CreateConversion(ctx context.Context, request CreateCheckoutRequest) (*CheckoutSession, error) {
	return s.create(ctx, "conversion", request)
}

func (s *CheckoutService) CreateDepositAndConvert(ctx context.Context, request CreateCheckoutRequest) (*CheckoutSession, error) {
	return s.create(ctx, "deposit_and_convert", request)
}

func (s *CheckoutService) create(ctx context.Context, checkoutType string, request CreateCheckoutRequest) (*CheckoutSession, error) {
	var out CheckoutSession
	err := s.client.doJSON(ctx, http.MethodPost, "/v1/checkout-sessions", nil,
		checkoutCreatePayload{Type: checkoutType, CreateCheckoutRequest: request}, &out, request.IdempotencyKey)
	return &out, err
}

func (s *CheckoutService) Get(ctx context.Context, sessionID string) (*CheckoutSession, error) {
	var out CheckoutSession
	err := s.client.doJSON(ctx, http.MethodGet, "/v1/checkout-sessions/"+url.PathEscape(sessionID), nil, nil, &out, "")
	return &out, err
}

type UsersService struct{ client *Client }

func (s *UsersService) CreateBindingSession(ctx context.Context, request UserBindingSessionRequest) (*UserBindingSession, error) {
	var out UserBindingSession
	err := s.client.doJSON(ctx, http.MethodPost, "/v1/user-binding-sessions", nil, request, &out, request.IdempotencyKey)
	return &out, err
}

func (s *UsersService) GetBalances(ctx context.Context, externalUserID string) (*UserBalances, error) {
	var out UserBalances
	err := s.client.doJSON(ctx, http.MethodGet, "/v1/users/"+url.PathEscape(externalUserID)+"/balances", nil, nil, &out, "")
	return &out, err
}

type ConversionsService struct{ client *Client }

func (s *ConversionsService) ListMarkets(ctx context.Context) ([]ConversionMarket, error) {
	var out ConversionMarketList
	err := s.client.doJSON(ctx, http.MethodGet, "/v1/conversion-markets", nil, nil, &out, "")
	return out.Items, err
}

func (s *ConversionsService) GetMarket(ctx context.Context, marketID string) (*ConversionMarket, error) {
	var out ConversionMarket
	err := s.client.doJSON(ctx, http.MethodGet, "/v1/conversion-markets/"+url.PathEscape(marketID), nil, nil, &out, "")
	return &out, err
}

func (s *ConversionsService) UpdatePrice(ctx context.Context, marketID string, request UpdateConversionPriceRequest) (*ConversionMarket, error) {
	var out ConversionMarket
	err := s.client.doJSON(ctx, http.MethodPut, "/v1/conversion-markets/"+url.PathEscape(marketID)+"/price", nil, request, &out, request.IdempotencyKey)
	return &out, err
}

func (s *ConversionsService) CreateQuote(ctx context.Context, request CreateConversionQuoteRequest) (*ConversionQuote, error) {
	var out ConversionQuote
	err := s.client.doJSON(ctx, http.MethodPost, "/v1/conversion-quotes", nil, request, &out, request.IdempotencyKey)
	return &out, err
}

type OperationsService struct{ client *Client }

func (s *OperationsService) Get(ctx context.Context, operationID string) (*Operation, error) {
	var out Operation
	err := s.client.doJSON(ctx, http.MethodGet, "/v1/operations/"+url.PathEscape(operationID), nil, nil, &out, "")
	return &out, err
}

func (s *OperationsService) List(ctx context.Context, params ListOperationsParams) (*OperationList, error) {
	query := url.Values{}
	if params.ExternalUserID != "" {
		query.Set("external_user_id", params.ExternalUserID)
	}
	if params.Type != "" {
		query.Set("type", params.Type)
	}
	if params.Status != "" {
		query.Set("status", params.Status)
	}
	if params.Cursor != "" {
		query.Set("cursor", params.Cursor)
	}
	if params.Limit > 0 {
		query.Set("limit", strconv.Itoa(params.Limit))
	}
	var out OperationList
	err := s.client.doJSON(ctx, http.MethodGet, "/v1/operations", query, nil, &out, "")
	return &out, err
}
