package resource

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/channels"
	"github.com/xautoop/dextri-pay-go/internal/transport"
)

type channelsService struct{ transport *transport.Client }

// NewChannels builds the channel resource implementation used by client.Client.
func NewChannels(client *transport.Client) *channelsService {
	return &channelsService{transport: client}
}

func (service *channelsService) List(ctx context.Context, params channels.ListParams) ([]channels.Channel, *api.Response, error) {
	query := url.Values{}
	if params.Flow != "" {
		if params.Flow != channels.FlowDeposit && params.Flow != channels.FlowWithdrawal {
			return nil, nil, &api.ValidationError{Field: "flow", Message: "must be deposit or withdrawal"}
		}
		query.Set("flow", string(params.Flow))
	}
	if value := strings.TrimSpace(params.SourceAsset); value != "" {
		query.Set("source_asset", strings.ToUpper(value))
	}
	var output []channels.Channel
	response, err := service.transport.Do(ctx, transport.Request{Method: http.MethodGet, Path: "/v1/channels", Query: query}, &output)
	return output, response, err
}
