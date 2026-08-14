package client

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/internal/transport"
	"github.com/xautoop/dextri-pay-go/payment"
)

type PaymentsService struct{ executor executor }

func newPaymentsService(executor executor) *PaymentsService {
	return &PaymentsService{executor: executor}
}

func (service *PaymentsService) Get(ctx context.Context, paymentID string) (*payment.Payment, *api.Response, error) {
	paymentID = strings.TrimSpace(paymentID)
	if paymentID == "" {
		return nil, nil, &api.ValidationError{Field: "payment_id", Message: "is required"}
	}
	var output payment.Payment
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodGet, Path: "/v1/payments/" + url.PathEscape(paymentID)}, &output)
	return &output, response, err
}

func (service *PaymentsService) Refund(ctx context.Context, paymentID string, request payment.RefundRequest, supplied ...RequestOption) (*payment.Refund, *api.Response, error) {
	paymentID = strings.TrimSpace(paymentID)
	if paymentID == "" {
		return nil, nil, &api.ValidationError{Field: "payment_id", Message: "is required"}
	}
	key, err := resolveIdempotencyKey(supplied...)
	if err != nil {
		return nil, nil, err
	}
	if err := request.Validate(); err != nil {
		return nil, nil, err
	}
	var output payment.Refund
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodPost, Path: "/v1/payments/" + url.PathEscape(paymentID) + "/refund", Body: request, IdempotencyKey: key}, &output)
	return &output, response, err
}
