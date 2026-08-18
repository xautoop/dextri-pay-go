package client

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/internal/transport"
	paytron "github.com/xautoop/dextri-pay-go/tron"
)

type TronService struct{ executor executor }

func newTronService(executor executor) *TronService { return &TronService{executor: executor} }

func (service *TronService) CreateDeposit(ctx context.Context, request paytron.CreateDepositRequest, supplied ...RequestOption) (*paytron.Deposit, *api.Response, error) {
	key, err := resolveIdempotencyKey(supplied...)
	if err != nil {
		return nil, nil, err
	}
	if err := request.Validate(); err != nil {
		return nil, nil, err
	}
	var output paytron.Deposit
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodPost, Path: "/v1/tron/deposits", Body: request, IdempotencyKey: key}, &output)
	return &output, response, err
}

func (service *TronService) GetDeposit(ctx context.Context, depositID string) (*paytron.Deposit, *api.Response, error) {
	depositID = strings.TrimSpace(depositID)
	if depositID == "" {
		return nil, nil, &api.ValidationError{Field: "deposit_id", Message: "is required"}
	}
	var output paytron.Deposit
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodGet, Path: "/v1/tron/deposits/" + url.PathEscape(depositID)}, &output)
	return &output, response, err
}

func (service *TronService) CreateWithdrawal(ctx context.Context, request paytron.CreateWithdrawalRequest, supplied ...RequestOption) (*paytron.Withdrawal, *api.Response, error) {
	key, err := resolveIdempotencyKey(supplied...)
	if err != nil {
		return nil, nil, err
	}
	if err := request.Validate(); err != nil {
		return nil, nil, err
	}
	var output paytron.Withdrawal
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodPost, Path: "/v1/tron/withdrawals", Body: request, IdempotencyKey: key}, &output)
	return &output, response, err
}

func (service *TronService) GetWithdrawal(ctx context.Context, withdrawalID string) (*paytron.Withdrawal, *api.Response, error) {
	withdrawalID = strings.TrimSpace(withdrawalID)
	if withdrawalID == "" {
		return nil, nil, &api.ValidationError{Field: "withdrawal_id", Message: "is required"}
	}
	var output paytron.Withdrawal
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodGet, Path: "/v1/tron/withdrawals/" + url.PathEscape(withdrawalID)}, &output)
	return &output, response, err
}

func (service *TronService) ListWithdrawals(ctx context.Context, status string, pageNo, pageSize int) (*paytron.WithdrawalList, *api.Response, error) {
	query := url.Values{}
	if status = strings.TrimSpace(status); status != "" {
		query.Set("status", status)
	}
	if pageNo > 0 {
		query.Set("page_no", strconv.Itoa(pageNo))
	}
	if pageSize > 0 {
		query.Set("page_size", strconv.Itoa(pageSize))
	}
	var output paytron.WithdrawalList
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodGet, Path: "/v1/tron/withdrawals", Query: query}, &output)
	return &output, response, err
}

func (service *TronService) ApproveWithdrawal(ctx context.Context, withdrawalID string, request paytron.ApproveRequest, supplied ...RequestOption) (*paytron.Withdrawal, *api.Response, error) {
	if err := request.Validate(); err != nil {
		return nil, nil, err
	}
	return service.review(ctx, withdrawalID, "approve", request, supplied...)
}

func (service *TronService) RejectWithdrawal(ctx context.Context, withdrawalID string, request paytron.RejectRequest, supplied ...RequestOption) (*paytron.Withdrawal, *api.Response, error) {
	if err := request.Validate(); err != nil {
		return nil, nil, err
	}
	return service.review(ctx, withdrawalID, "reject", request, supplied...)
}

func (service *TronService) review(ctx context.Context, withdrawalID, action string, body any, supplied ...RequestOption) (*paytron.Withdrawal, *api.Response, error) {
	withdrawalID = strings.TrimSpace(withdrawalID)
	if withdrawalID == "" {
		return nil, nil, &api.ValidationError{Field: "withdrawal_id", Message: "is required"}
	}
	key, err := resolveIdempotencyKey(supplied...)
	if err != nil {
		return nil, nil, err
	}
	var output paytron.Withdrawal
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodPost, Path: "/v1/tron/withdrawals/" + url.PathEscape(withdrawalID) + "/" + action, Body: body, IdempotencyKey: key}, &output)
	return &output, response, err
}
