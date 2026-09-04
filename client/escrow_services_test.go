package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/xautoop/dextri-pay-go/burn"
	"github.com/xautoop/dextri-pay-go/escrow"
	"github.com/xautoop/dextri-pay-go/hold"
	"github.com/xautoop/dextri-pay-go/internal/auth"
	"github.com/xautoop/dextri-pay-go/money"
)

func TestEscrowServiceContracts(t *testing.T) {
	client := testClient(t, func(request *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(request.Body)
		if request.Method == http.MethodPost && request.Header.Get(auth.HeaderIdempotency) == "" {
			t.Fatal("mutation missing idempotency key")
		}
		var payload map[string]any
		if len(body) != 0 {
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatal(err)
			}
		}
		switch request.URL.Path {
		case "/pay/v1/app-accounts/weekly_pool/balances/DXS":
			return jsonResponse(200, `{"account_key":"weekly_pool","asset":"DXS","total":"50","available":"50","locked":"0","frozen":"0","enabled":true,"capabilities":["receive","payout"]}`, nil), nil
		case "/pay/v1/holds":
			if payload["external_user_id"] != "user-1" || payload["amount"] != "500.00" {
				t.Fatalf("hold payload=%#v", payload)
			}
			return jsonResponse(201, `{"hold_id":"h1","operation_id":"op1","status":"held","amount":"500.00"}`, nil), nil
		case "/pay/v1/holds/h1":
			return jsonResponse(200, `{"hold_id":"h1","operation_id":"op1","status":"held","amount":"500.00"}`, nil), nil
		case "/pay/v1/holds/h1/release":
			return jsonResponse(200, `{"hold_id":"h1","operation_id":"op2","status":"released","amount":"500.00"}`, nil), nil
		case "/pay/v1/escrows/commit":
			return jsonResponse(201, `{"escrow_id":"e1","operation_id":"op3","status":"committed","asset":"DXS","amount":"1000.00","amount_atomic":"1000000000","hold_ids":["h1","h2"]}`, nil), nil
		case "/pay/v1/escrow-settlements":
			if payload["total_amount_atomic"] != "1000000000" {
				t.Fatalf("settlement payload=%#v", payload)
			}
			return jsonResponse(201, `{"settlement_id":"s1","operation_id":"op4","escrow_id":"e1","status":"consumed","asset":"DXS","total_amount":"1000.00","total_amount_atomic":"1000000000"}`, nil), nil
		case "/pay/v1/burns":
			if payload["source_account_key"] != "burn_pending" || payload["amount_atomic"] != "50000000" {
				t.Fatalf("burn payload=%#v", payload)
			}
			return jsonResponse(201, `{"burn_id":"b1","operation_id":"op5","source_account_key":"burn_pending","status":"processing","asset":"DXS","amount":"50.00","amount_atomic":"50000000"}`, nil), nil
		case "/pay/v1/burns/b1":
			return jsonResponse(200, `{"burn_id":"b1","operation_id":"op5","source_account_key":"burn_pending","status":"succeeded","asset":"DXS","amount":"50.00","amount_atomic":"50000000"}`, nil), nil
		default:
			t.Fatalf("unexpected request %s %s", request.Method, request.URL.Path)
			return nil, nil
		}
	})
	ctx := context.Background()
	balance, _, err := client.Accounts.GetBalance(ctx, "weekly_pool", "DXS")
	if err != nil || !balance.Enabled || balance.Available != "50" {
		t.Fatalf("balance=%#v err=%v", balance, err)
	}
	created, _, err := client.Holds.Create(ctx, hold.CreateRequest{
		ExternalUserID: "user-1", ClientReferenceID: "order-1", Asset: "DXS", Amount: "500.00",
		Purpose: "game_stake", ExpiresAt: time.Now().Add(time.Hour),
	}, WithIdempotencyKey("hold-0001"))
	if err != nil || created.Status != hold.StatusHeld {
		t.Fatalf("hold=%#v err=%v", created, err)
	}
	if _, _, err := client.Holds.Get(ctx, "h1"); err != nil {
		t.Fatal(err)
	}
	released, _, err := client.Holds.Release(ctx, "h1", hold.ReleaseRequest{Reason: "expired"}, WithIdempotencyKey("release-0001"))
	if err != nil || released.Status != hold.StatusReleased {
		t.Fatalf("release=%#v err=%v", released, err)
	}
	committed, _, err := client.Escrows.Commit(ctx, escrow.CommitRequest{ClientReferenceID: "order-1", HoldIDs: []string{"h1", "h2"}}, WithIdempotencyKey("commit-0001"))
	if err != nil || committed.Status != escrow.StatusCommitted {
		t.Fatalf("escrow=%#v err=%v", committed, err)
	}
	settled, _, err := client.Settlements.Create(ctx, validSettlementRequest(), WithIdempotencyKey("settle-0001"))
	if err != nil || settled.SettlementID != "s1" {
		t.Fatalf("settlement=%#v err=%v", settled, err)
	}
	burned, _, err := client.Burns.Create(ctx, burn.CreateRequest{
		SourceAccountKey: "burn_pending", ClientReferenceID: "burn-week-1", Asset: "DXS",
		Amount: "50.00", AmountAtomic: "50000000", Reason: "scheduled_burn",
	}, WithIdempotencyKey("burn-0001"))
	if err != nil || burned.BurnID != "b1" {
		t.Fatalf("burn=%#v err=%v", burned, err)
	}
	if burned, _, err = client.Burns.Get(ctx, "b1"); err != nil || burned.Status != burn.StatusSucceeded {
		t.Fatalf("burn=%#v err=%v", burned, err)
	}
}

func validSettlementRequest() escrow.SettlementRequest {
	return escrow.SettlementRequest{
		EscrowID: "e1", ClientReferenceID: "settle-1", Asset: "DXS", TotalAmount: "1000.00", TotalAmountAtomic: "1000000000",
		Allocations: []escrow.Allocation{
			{TargetType: escrow.TargetExternalUser, TargetID: "winner", Amount: money.Decimal("700.00"), AmountAtomic: "700000000"},
			{TargetType: escrow.TargetAppAccount, TargetID: "revenue", Amount: money.Decimal("200.00"), AmountAtomic: "200000000"},
			{TargetType: escrow.TargetAppAccount, TargetID: "pool", Amount: money.Decimal("50.00"), AmountAtomic: "50000000"},
			{TargetType: escrow.TargetAppAccount, TargetID: "burn", Amount: money.Decimal("50.00"), AmountAtomic: "50000000"},
		},
	}
}
