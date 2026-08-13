package client

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/channels"
	"github.com/xautoop/dextri-pay-go/internal/transport"
)

// ChannelsService discovers routes authorized for the authenticated App.
type ChannelsService struct {
	executor executor
}

func newChannelsService(executor executor) *ChannelsService {
	return &ChannelsService{executor: executor}
}

// List returns the healthy, authorized channels matching params.
func (service *ChannelsService) List(ctx context.Context, params channels.ListParams) ([]channels.Channel, *api.Response, error) {
	if err := params.Validate(); err != nil {
		return nil, nil, err
	}
	query := url.Values{}
	if params.Flow != "" {
		query.Set("flow", string(params.Flow))
	}
	if value := strings.TrimSpace(params.SourceAsset); value != "" {
		query.Set("source_asset", strings.ToUpper(value))
	}
	var output []channels.Channel
	response, err := service.executor.Do(ctx, transport.Request{Method: http.MethodGet, Path: "/v1/channels", Query: query}, &output)
	return output, response, err
}
