package checkout

import "testing"

type checkoutRequest interface {
	Validate() error
}

func TestCreateRequestsValidate(t *testing.T) {
	valid := []checkoutRequest{
		CreateDepositRequest{ExternalUserID: "user", SourceAsset: "USDT", TargetAsset: "USDC", Amount: "1"},
		CreateWithdrawalRequest{ExternalUserID: "user", SourceAsset: "USDC", TargetAsset: "USDT", Amount: "1"},
		CreateConversionRequest{ExternalUserID: "user", SourceAsset: "ASSET-A", TargetAsset: "ASSET-B", Amount: "1"},
		CreateDepositAndConvertRequest{ExternalUserID: "user", SourceAsset: "USDT", TargetAsset: "ASSET-A", Amount: "1"},
	}
	for _, request := range valid {
		if err := request.Validate(); err != nil {
			t.Fatalf("Validate(%T) error = %v", request, err)
		}
	}

	invalid := []checkoutRequest{
		CreateDepositRequest{SourceAsset: "USDT", TargetAsset: "USDC", Amount: "1"},
		CreateWithdrawalRequest{ExternalUserID: "user", TargetAsset: "USDT", Amount: "1"},
		CreateConversionRequest{ExternalUserID: "user", SourceAsset: "ASSET-A", TargetAsset: "ASSET-B", Amount: "0"},
	}
	for _, request := range invalid {
		if err := request.Validate(); err == nil {
			t.Fatalf("invalid %T accepted", request)
		}
	}
}
