package resource

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/conversion"
	"github.com/xautoop/dextri-pay-go/internal/transport"
)

type conversionService struct{ transport *transport.Client }

// NewConversions builds the conversion resource implementation used by client.Client.
func NewConversions(client *transport.Client) *conversionService {
	return &conversionService{transport: client}
}

func (service *conversionService) ListMarkets(ctx context.Context) (*conversion.MarketList, *api.Response, error) {
	var output conversion.MarketList
	response, err := service.transport.Do(ctx, transport.Request{Method: http.MethodGet, Path: "/v1/conversion-markets"}, &output)
	return &output, response, err
}

func (service *conversionService) GetMarket(ctx context.Context, id string) (*conversion.Market, *api.Response, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, nil, &api.ValidationError{Field: "market_id", Message: "is required"}
	}
	var output conversion.Market
	response, err := service.transport.Do(ctx, transport.Request{Method: http.MethodGet, Path: "/v1/conversion-markets/" + url.PathEscape(id)}, &output)
	return &output, response, err
}

func (service *conversionService) UpdatePrice(ctx context.Context, id string, request conversion.UpdatePriceRequest, idempotencyKey string) (*conversion.Market, *api.Response, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, nil, &api.ValidationError{Field: "market_id", Message: "is required"}
	}
	if err := request.Validate(); err != nil {
		return nil, nil, &api.ValidationError{Field: "conversion_price", Message: err.Error()}
	}
	var output conversion.Market
	response, err := service.transport.Do(ctx, transport.Request{Method: http.MethodPut, Path: "/v1/conversion-markets/" + url.PathEscape(id) + "/price", Body: request, IdempotencyKey: idempotencyKey}, &output)
	return &output, response, err
}

func (service *conversionService) CreateQuote(ctx context.Context, request conversion.CreateQuoteRequest, idempotencyKey string) (*conversion.Quote, *api.Response, error) {
	request.ExternalUserID = strings.TrimSpace(request.ExternalUserID)
	request.Owner = strings.TrimSpace(request.Owner)
	request.MarketID = strings.TrimSpace(request.MarketID)
	if request.ExternalUserID == "" {
		return nil, nil, &api.ValidationError{Field: "external_user_id", Message: "is required"}
	}
	if request.MarketID == "" {
		return nil, nil, &api.ValidationError{Field: "market_id", Message: "is required"}
	}
	if !request.Side.Valid() {
		return nil, nil, &api.ValidationError{Field: "side", Message: "must be sell_base or buy_base"}
	}
	if err := request.InputAmount.ValidatePositive(); err != nil {
		return nil, nil, &api.ValidationError{Field: "input_amount", Message: err.Error()}
	}
	var output conversion.Quote
	response, err := service.transport.Do(ctx, transport.Request{Method: http.MethodPost, Path: "/v1/conversion-quotes", Body: request, IdempotencyKey: idempotencyKey}, &output)
	return &output, response, err
}
