package client

import (
	"context"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/internal/transport"
)

// executor is the private boundary between public resource services and HTTP.
// Keeping the interface here makes each service directly testable without an
// additional forwarding implementation layer.
type executor interface {
	Do(context.Context, transport.Request, any) (*api.Response, error)
}
