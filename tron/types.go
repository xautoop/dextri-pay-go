// Package tron defines App-facing TRON USDT deposit and manually reviewed
// withdrawal contracts. Pay verifies chain evidence; it never exposes keys or
// broadcasts phase-one withdrawals.
package tron

import (
	"encoding/hex"
	"strings"
	"time"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/money"
)

type DepositMode string

const (
	DepositModeMemo    DepositMode = "tron_usdt_memo"
	DepositModeManaged DepositMode = "tron_usdt_managed"
)

type CreateDepositRequest struct {
	ExternalUserID    string       `json:"external_user_id"`
	ClientReferenceID string       `json:"client_reference_id"`
	Mode              DepositMode  `json:"mode"`
	Metadata          api.Metadata `json:"metadata,omitempty"`
}

func (request CreateDepositRequest) Validate() error {
	if strings.TrimSpace(request.ExternalUserID) == "" {
		return &api.ValidationError{Field: "external_user_id", Message: "is required"}
	}
	if strings.TrimSpace(request.ClientReferenceID) == "" {
		return &api.ValidationError{Field: "client_reference_id", Message: "is required"}
	}
	if request.Mode != DepositModeMemo && request.Mode != DepositModeManaged {
		return &api.ValidationError{Field: "mode", Message: "must be tron_usdt_memo or tron_usdt_managed"}
	}
	return nil
}

type Deposit struct {
	DepositID             string        `json:"deposit_id"`
	OperationID           string        `json:"operation_id"`
	ExternalUserID        string        `json:"external_user_id"`
	ClientReferenceID     string        `json:"client_reference_id,omitempty"`
	Mode                  DepositMode   `json:"mode"`
	Network               string        `json:"network"`
	Asset                 string        `json:"asset"`
	TokenAddress          string        `json:"token_address"`
	TokenDecimals         uint32        `json:"token_decimals"`
	DepositAddress        string        `json:"deposit_address"`
	Memo                  string        `json:"memo,omitempty"`
	Status                string        `json:"status"`
	Amount                money.Decimal `json:"amount,omitempty"`
	AmountAtomic          string        `json:"amount_atomic,omitempty"`
	TxHash                string        `json:"tx_hash,omitempty"`
	Confirmations         uint64        `json:"confirmations,omitempty"`
	RequiredConfirmations uint64        `json:"required_confirmations"`
	ExpiresAt             time.Time     `json:"expires_at"`
	CreatedAt             time.Time     `json:"created_at"`
	CompletedAt           *time.Time    `json:"completed_at,omitempty"`
}

type CreateWithdrawalRequest struct {
	ExternalUserID     string        `json:"external_user_id"`
	ClientReferenceID  string        `json:"client_reference_id"`
	Amount             money.Decimal `json:"amount"`
	DestinationAddress string        `json:"destination_address"`
	Metadata           api.Metadata  `json:"metadata,omitempty"`
}

func (request CreateWithdrawalRequest) Validate() error {
	if strings.TrimSpace(request.ExternalUserID) == "" {
		return &api.ValidationError{Field: "external_user_id", Message: "is required"}
	}
	if strings.TrimSpace(request.ClientReferenceID) == "" {
		return &api.ValidationError{Field: "client_reference_id", Message: "is required"}
	}
	if err := request.Amount.ValidatePositive(); err != nil {
		return &api.ValidationError{Field: "amount", Message: err.Error()}
	}
	if strings.TrimSpace(request.DestinationAddress) == "" {
		return &api.ValidationError{Field: "destination_address", Message: "is required"}
	}
	return nil
}

type Withdrawal struct {
	WithdrawalID          string        `json:"withdrawal_id"`
	OperationID           string        `json:"operation_id"`
	ExternalUserID        string        `json:"external_user_id"`
	ClientReferenceID     string        `json:"client_reference_id,omitempty"`
	Network               string        `json:"network"`
	Asset                 string        `json:"asset"`
	TokenAddress          string        `json:"token_address"`
	TokenDecimals         uint32        `json:"token_decimals"`
	SourceAddress         string        `json:"source_address"`
	DestinationAddress    string        `json:"destination_address"`
	Amount                money.Decimal `json:"amount"`
	AmountAtomic          string        `json:"amount_atomic"`
	Status                string        `json:"status"`
	TxHash                string        `json:"tx_hash,omitempty"`
	Confirmations         uint64        `json:"confirmations,omitempty"`
	RequiredConfirmations uint64        `json:"required_confirmations"`
	ReviewActor           string        `json:"review_actor,omitempty"`
	ReviewReason          string        `json:"review_reason,omitempty"`
	CreatedAt             time.Time     `json:"created_at"`
	ReviewedAt            *time.Time    `json:"reviewed_at,omitempty"`
	CompletedAt           *time.Time    `json:"completed_at,omitempty"`
}

type WithdrawalList struct {
	Items []Withdrawal `json:"items"`
	Total int64        `json:"total"`
}

type ReviewRequest struct {
	Actor string `json:"actor"`
}

func (request ReviewRequest) Validate() error {
	if strings.TrimSpace(request.Actor) == "" {
		return &api.ValidationError{Field: "actor", Message: "is required"}
	}
	return nil
}

type RejectRequest struct {
	Actor  string `json:"actor"`
	Reason string `json:"reason"`
}

func (request RejectRequest) Validate() error {
	if strings.TrimSpace(request.Actor) == "" {
		return &api.ValidationError{Field: "actor", Message: "is required"}
	}
	if strings.TrimSpace(request.Reason) == "" {
		return &api.ValidationError{Field: "reason", Message: "is required"}
	}
	return nil
}

type SubmitTransferRequest struct {
	Actor  string `json:"actor"`
	TxHash string `json:"tx_hash"`
}

func (request SubmitTransferRequest) Validate() error {
	if strings.TrimSpace(request.Actor) == "" {
		return &api.ValidationError{Field: "actor", Message: "is required"}
	}
	txHash := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(request.TxHash), "0x"), "0X")
	if len(txHash) != 64 {
		return &api.ValidationError{Field: "tx_hash", Message: "must be a 64-character hexadecimal TRON transaction hash"}
	}
	if _, err := hex.DecodeString(txHash); err != nil {
		return &api.ValidationError{Field: "tx_hash", Message: "must be a 64-character hexadecimal TRON transaction hash"}
	}
	return nil
}
