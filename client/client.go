package client

import (
	"fmt"
	"strings"

	"github.com/xautoop/dextri-pay-go/internal/auth"
	"github.com/xautoop/dextri-pay-go/internal/transport"
)

// Client is safe for concurrent use after construction. Its service fields
// must not be reassigned while requests are in flight.
type Client struct {
	baseURL string
	appID   string
	keyID   string

	// Channels discovers routes authorized for the authenticated App.
	Channels *ChannelsService
	// Checkout creates and reads durable Hosted Checkout sessions.
	Checkout *CheckoutService
	// Users creates wallet bindings and reads live user balances.
	Users *UsersService
	// Conversions manages dynamic App-priced conversion markets and quotes.
	Conversions *ConversionsService
	// Operations reads durable App-visible business operations.
	Operations *OperationsService
	// Payments reads and refunds App Commerce payments.
	Payments *PaymentsService
	// Payouts creates and reads optional App-funded disbursements.
	Payouts *PayoutsService
	// Accounts reads registered App-account balances and capabilities.
	Accounts *AccountsService
	// Holds creates, reads and releases generic balance reservations.
	Holds *HoldsService
	// Escrows atomically commits compatible Holds.
	Escrows *EscrowsService
	// Settlements atomically distributes committed Escrow funds.
	Settlements *SettlementsService
	// Tron creates USDT deposit instructions and manually reviewed withdrawals.
	Tron *TronService
}

// New constructs a concurrency-safe Pay client from explicit credentials and options.
func New(config Config, supplied ...Option) (*Client, error) {
	options := defaultOptions()
	for _, option := range supplied {
		if option != nil {
			if err := option(&options); err != nil {
				return nil, err
			}
		}
	}
	userAgent := "dextri-pay-go"
	if strings.TrimSpace(config.UserAgent) != "" {
		userAgent += " " + strings.TrimSpace(config.UserAgent)
	}
	transportClient, err := transport.New(transport.Config{
		BaseURL: config.BaseURL,
		Credentials: auth.Credentials{
			AppID:  config.Credentials.AppID,
			KeyID:  config.Credentials.KeyID,
			Secret: config.Credentials.Secret,
		},
		HTTPClient:        options.httpClient,
		RetryPolicy:       options.retry,
		AllowInsecureHTTP: options.allowInsecureHTTP,
		UserAgent:         userAgent,
		Observer:          options.observer,
		Now:               options.now,
		Nonce:             options.nonce,
	})
	if err != nil {
		return nil, err
	}
	return &Client{
		baseURL:     strings.TrimSpace(config.BaseURL),
		appID:       strings.TrimSpace(config.Credentials.AppID),
		keyID:       strings.TrimSpace(config.Credentials.KeyID),
		Channels:    newChannelsService(transportClient),
		Checkout:    newCheckoutService(transportClient),
		Users:       newUsersService(transportClient),
		Conversions: newConversionsService(transportClient),
		Operations:  newOperationsService(transportClient),
		Payments:    newPaymentsService(transportClient),
		Payouts:     newPayoutsService(transportClient),
		Accounts:    newAccountsService(transportClient),
		Holds:       newHoldsService(transportClient),
		Escrows:     newEscrowsService(transportClient),
		Settlements: newSettlementsService(transportClient),
		Tron:        newTronService(transportClient),
	}, nil
}

func (client *Client) String() string {
	if client == nil {
		return "client.Client<nil>"
	}
	return fmt.Sprintf("client.Client{base_url:%q, app_id:%q, key_id:%q, secret:[REDACTED]}", client.baseURL, client.appID, client.keyID)
}

// GoString returns a redacted diagnostic representation of the client.
func (client *Client) GoString() string { return client.String() }
