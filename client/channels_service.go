package client

import (
	"context"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/channels"
)

type channelsBackend interface {
	List(context.Context, channels.ListParams) ([]channels.Channel, *api.Response, error)
}

// ChannelsService discovers routes authorized for the authenticated App.
type ChannelsService struct {
	backend channelsBackend
}

func newChannelsService(backend channelsBackend) *ChannelsService {
	return &ChannelsService{backend: backend}
}

// List returns the healthy, authorized channels matching params.
func (service *ChannelsService) List(ctx context.Context, params channels.ListParams) ([]channels.Channel, *api.Response, error) {
	return service.backend.List(ctx, params)
}
