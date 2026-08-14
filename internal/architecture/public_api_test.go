package architecture

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/channels"
	"github.com/xautoop/dextri-pay-go/checkout"
	"github.com/xautoop/dextri-pay-go/client"
	"github.com/xautoop/dextri-pay-go/conversion"
	"github.com/xautoop/dextri-pay-go/operation"
	"github.com/xautoop/dextri-pay-go/payment"
	"github.com/xautoop/dextri-pay-go/payout"
	"github.com/xautoop/dextri-pay-go/users"
	"github.com/xautoop/dextri-pay-go/webhook"
)

type (
	channelsListMethod func(
		*client.ChannelsService,
		context.Context,
		channels.ListParams,
	) ([]channels.Channel, *api.Response, error)
	checkoutDepositMethod func(
		*client.CheckoutService,
		context.Context,
		checkout.CreateDepositRequest,
		...client.RequestOption,
	) (*checkout.Session, *api.Response, error)
	checkoutWithdrawalMethod func(
		*client.CheckoutService,
		context.Context,
		checkout.CreateWithdrawalRequest,
		...client.RequestOption,
	) (*checkout.Session, *api.Response, error)
	checkoutConversionMethod func(
		*client.CheckoutService,
		context.Context,
		checkout.CreateConversionRequest,
		...client.RequestOption,
	) (*checkout.Session, *api.Response, error)
	checkoutDepositAndConvertMethod func(
		*client.CheckoutService,
		context.Context,
		checkout.CreateDepositAndConvertRequest,
		...client.RequestOption,
	) (*checkout.Session, *api.Response, error)
	checkoutGetMethod func(
		*client.CheckoutService,
		context.Context,
		string,
	) (*checkout.Session, *api.Response, error)
	checkoutPaymentMethod    func(*client.CheckoutService, context.Context, checkout.CreatePaymentRequest, ...client.RequestOption) (*checkout.Session, *api.Response, error)
	paymentGetMethod         func(*client.PaymentsService, context.Context, string) (*payment.Payment, *api.Response, error)
	paymentRefundMethod      func(*client.PaymentsService, context.Context, string, payment.RefundRequest, ...client.RequestOption) (*payment.Refund, *api.Response, error)
	payoutCreateMethod       func(*client.PayoutsService, context.Context, payout.CreateRequest, ...client.RequestOption) (*payout.Payout, *api.Response, error)
	payoutGetMethod          func(*client.PayoutsService, context.Context, string) (*payout.Payout, *api.Response, error)
	usersCreateBindingMethod func(
		*client.UsersService,
		context.Context,
		users.CreateBindingSessionRequest,
		...client.RequestOption,
	) (*users.BindingSession, *api.Response, error)
	usersGetBalancesMethod func(
		*client.UsersService,
		context.Context,
		string,
	) (*users.Balances, *api.Response, error)
	conversionsListMarketsMethod func(
		*client.ConversionsService,
		context.Context,
	) (*conversion.MarketList, *api.Response, error)
	conversionsGetMarketMethod func(
		*client.ConversionsService,
		context.Context,
		string,
	) (*conversion.Market, *api.Response, error)
	conversionsUpdatePriceMethod func(
		*client.ConversionsService,
		context.Context,
		string,
		conversion.UpdatePriceRequest,
		...client.RequestOption,
	) (*conversion.Market, *api.Response, error)
	conversionsCreateQuoteMethod func(
		*client.ConversionsService,
		context.Context,
		conversion.CreateQuoteRequest,
		...client.RequestOption,
	) (*conversion.Quote, *api.Response, error)
	operationsGetMethod func(
		*client.OperationsService,
		context.Context,
		string,
	) (*operation.Operation, *api.Response, error)
	operationsListMethod func(
		*client.OperationsService,
		context.Context,
		operation.ListParams,
	) (*operation.List, *api.Response, error)
)

// These compile-time assignments pin the first-release SDK entry points. A
// deliberate breaking change must update this manifest and the changelog.
var (
	_ func(client.Config, ...client.Option) (*client.Client, error) = client.New
	_ func(*http.Client) client.Option                              = client.WithHTTPClient
	_ func(int, time.Duration, time.Duration) client.Option         = client.WithRetryPolicy
	_ func(string) client.RequestOption                             = client.WithIdempotencyKey

	_ channelsListMethod              = (*client.ChannelsService).List
	_ checkoutDepositMethod           = (*client.CheckoutService).CreateDeposit
	_ checkoutWithdrawalMethod        = (*client.CheckoutService).CreateWithdrawal
	_ checkoutConversionMethod        = (*client.CheckoutService).CreateConversion
	_ checkoutDepositAndConvertMethod = (*client.CheckoutService).CreateDepositAndConvert
	_ checkoutGetMethod               = (*client.CheckoutService).Get
	_ checkoutPaymentMethod           = (*client.CheckoutService).CreatePayment
	_ paymentGetMethod                = (*client.PaymentsService).Get
	_ paymentRefundMethod             = (*client.PaymentsService).Refund
	_ payoutCreateMethod              = (*client.PayoutsService).Create
	_ payoutGetMethod                 = (*client.PayoutsService).Get
	_ usersCreateBindingMethod        = (*client.UsersService).CreateBindingSession
	_ usersGetBalancesMethod          = (*client.UsersService).GetBalances
	_ conversionsListMarketsMethod    = (*client.ConversionsService).ListMarkets
	_ conversionsGetMarketMethod      = (*client.ConversionsService).GetMarket
	_ conversionsUpdatePriceMethod    = (*client.ConversionsService).UpdatePrice
	_ conversionsCreateQuoteMethod    = (*client.ConversionsService).CreateQuote
	_ operationsGetMethod             = (*client.OperationsService).Get
	_ operationsListMethod            = (*client.OperationsService).List

	_ func(string, http.Header, []byte) (*webhook.Delivery, error)                           = webhook.Verify
	_ func(string, http.Header, []byte, time.Time, time.Duration) (*webhook.Delivery, error) = webhook.VerifyAt
	_ func(webhook.Verifier, http.Header, []byte) (*webhook.Delivery, error)                 = webhook.Verifier.Verify
)

func assertClientServiceFields(pay *client.Client) {
	_ = pay.Channels
	_ = pay.Checkout
	_ = pay.Users
	_ = pay.Conversions
	_ = pay.Operations
	_ = pay.Payments
	_ = pay.Payouts
}

func assertErrorAndResponseFields(apiError *api.APIError, requestError *api.RequestError, validationError *api.ValidationError, response *api.Response) {
	var _ int = apiError.StatusCode
	var _ string = apiError.Code
	var _ string = apiError.Message
	var _ string = apiError.RequestID
	var _ string = apiError.IdempotencyKey
	var _ json.RawMessage = apiError.Details

	var _ string = requestError.RequestID
	var _ string = requestError.IdempotencyKey
	var _ int = requestError.Attempts
	var _ error = requestError.Err

	var _ string = validationError.Field
	var _ string = validationError.Message

	var _ int = response.StatusCode
	var _ string = response.RequestID
	var _ string = response.IdempotencyKey
	var _ int = response.Attempts
	var _ http.Header = response.Headers
}

func TestPublicFieldsCompile(_ *testing.T) {
	assertClientServiceFields(&client.Client{})
	assertErrorAndResponseFields(&api.APIError{}, &api.RequestError{}, &api.ValidationError{}, &api.Response{})
}
