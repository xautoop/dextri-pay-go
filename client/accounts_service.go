package client

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/xautoop/dextri-pay-go/account"
	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/internal/transport"
)

// AccountsService reads registered App settlement accounts.
type AccountsService struct{ executor executor }

func newAccountsService(executor executor) *AccountsService {
	return &AccountsService{executor: executor}
}

// GetBalance returns the authoritative balance and capabilities for one App account asset.
func (service *AccountsService) GetBalance(ctx context.Context, accountKey, asset string) (*account.Balance, *api.Response, error) {
	accountKey = strings.TrimSpace(accountKey)
	asset = strings.TrimSpace(asset)
	if accountKey == "" {
		return nil, nil, &api.ValidationError{Field: "account_key", Message: "is required"}
	}
	if asset == "" {
		return nil, nil, &api.ValidationError{Field: "asset", Message: "is required"}
	}
	var output account.Balance
	path := "/v1/app-accounts/" + url.PathEscape(accountKey) + "/balances/" + url.PathEscape(asset)
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodGet, Path: path}, &output)
	return &output, response, err
}
