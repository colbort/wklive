package adminlogic

import (
	"context"
	"errors"
	"testing"
	"time"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

func TestInsuranceInventoryExitWritePathsDisabledByDefault(t *testing.T) {
	serviceCtx := &svc.ServiceContext{}
	tests := []struct {
		name string
		call func() *option.GetInsuranceInventoryExitResp
	}{
		{name: "create", call: func() *option.GetInsuranceInventoryExitResp {
			resp, err := NewCreateInsuranceInventoryExitLogic(context.Background(), serviceCtx).
				CreateInsuranceInventoryExit(&option.CreateInsuranceInventoryExitReq{})
			if err != nil {
				t.Fatalf("create returned transport error: %v", err)
			}
			return resp
		}},
		{name: "review", call: func() *option.GetInsuranceInventoryExitResp {
			resp, err := NewReviewInsuranceInventoryExitLogic(context.Background(), serviceCtx).
				ReviewInsuranceInventoryExit(&option.ReviewInsuranceInventoryExitReq{})
			if err != nil {
				t.Fatalf("review returned transport error: %v", err)
			}
			return resp
		}},
		{name: "execute", call: func() *option.GetInsuranceInventoryExitResp {
			resp, err := NewExecuteInsuranceInventoryExitLogic(context.Background(), serviceCtx).
				ExecuteInsuranceInventoryExit(&option.ExecuteInsuranceInventoryExitReq{})
			if err != nil {
				t.Fatalf("execute returned transport error: %v", err)
			}
			return resp
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp := test.call()
			if resp == nil || resp.Base == nil || resp.Base.Code == 200 ||
				resp.Base.Msg != errInsuranceInventoryExitDisabled.Error() {
				t.Fatalf("expected disabled response, got %+v", resp)
			}
		})
	}
}

func validInsuranceInventoryExitLimitContext() *svc.ServiceContext {
	serviceCtx := &svc.ServiceContext{}
	serviceCtx.Config.InsuranceInventoryExit.Enabled = true
	serviceCtx.Config.InsuranceInventoryExit.MaxQuantityPerRequest = "10"
	serviceCtx.Config.InsuranceInventoryExit.MaxPremiumPerRequest = "1000"
	serviceCtx.Config.InsuranceInventoryExit.MaxDailyQuantity = "20"
	serviceCtx.Config.InsuranceInventoryExit.MaxMarkDeviationRatio = "0.10"
	serviceCtx.Config.InsuranceInventoryExit.MinOrderBookQuantity = "2"
	return serviceCtx
}

func TestInsuranceInventoryExitRuntimeLimitsFailClosed(t *testing.T) {
	serviceCtx := validInsuranceInventoryExitLimitContext()
	limits, err := insuranceInventoryExitRuntimeLimits(serviceCtx)
	if err != nil || !limits.maxDailyQuantity.Equal(decimal.NewFromInt(20)) {
		t.Fatalf("valid limits=%+v err=%v", limits, err)
	}
	serviceCtx.Config.InsuranceInventoryExit.MaxDailyQuantity = "0"
	if _, err = insuranceInventoryExitRuntimeLimits(serviceCtx); !errors.Is(err, errInsuranceInventoryExitLimits) {
		t.Fatalf("expected invalid zero daily limit, got %v", err)
	}
	serviceCtx.Config.InsuranceInventoryExit.MaxDailyQuantity = "5"
	if _, err = insuranceInventoryExitRuntimeLimits(serviceCtx); !errors.Is(err, errInsuranceInventoryExitLimits) {
		t.Fatalf("expected daily limit below per-request limit to fail, got %v", err)
	}
}

func TestValidateInsuranceInventoryExitRuntimeLimits(t *testing.T) {
	limits, err := insuranceInventoryExitRuntimeLimits(validInsuranceInventoryExitLimitContext())
	if err != nil {
		t.Fatal(err)
	}
	contract := &models.TOptionContract{Multiplier: decimal.NewFromInt(1)}
	market := &models.TOptionMarket{MarkPrice: decimal.NewFromInt(100)}
	if err = validateInsuranceInventoryExitRuntimeLimits(
		contract, market, decimal.NewFromInt(5), decimal.NewFromInt(105), limits,
	); err != nil {
		t.Fatalf("valid runtime limits rejected: %v", err)
	}
	if err = validateInsuranceInventoryExitRuntimeLimits(
		contract, market, decimal.NewFromInt(11), decimal.NewFromInt(100), limits,
	); !errors.Is(err, errInsuranceInventoryExitLimits) {
		t.Fatalf("expected quantity rejection, got %v", err)
	}
	if err = validateInsuranceInventoryExitRuntimeLimits(
		contract, market, decimal.NewFromInt(10), decimal.NewFromInt(101), limits,
	); !errors.Is(err, errInsuranceInventoryExitLimits) {
		t.Fatalf("expected premium rejection, got %v", err)
	}
	if err = validateInsuranceInventoryExitRuntimeLimits(
		contract, market, decimal.NewFromInt(1), decimal.NewFromInt(111), limits,
	); !errors.Is(err, errInsuranceInventoryExitLimits) {
		t.Fatalf("expected mark-deviation rejection, got %v", err)
	}
}

func TestValidateInsuranceInventoryExitOrderBookDepth(t *testing.T) {
	limits, err := insuranceInventoryExitRuntimeLimits(validInsuranceInventoryExitLimitContext())
	if err != nil {
		t.Fatal(err)
	}
	orders := []*models.TOptionOrder{
		{UnfilledQty: decimal.RequireFromString("0.5")},
		{UnfilledQty: decimal.RequireFromString("1.5")},
	}
	if err = validateInsuranceInventoryExitOrderBookDepth(orders, limits); err != nil {
		t.Fatalf("exact minimum depth rejected: %v", err)
	}
	orders[1].UnfilledQty = decimal.RequireFromString("1.499999999999999999")
	if err = validateInsuranceInventoryExitOrderBookDepth(orders, limits); !errors.Is(err, errInsuranceInventoryExitLimits) {
		t.Fatalf("below-minimum depth accepted: %v", err)
	}
}

func TestInsuranceInventoryExitUTCDayStart(t *testing.T) {
	now := time.Date(2026, time.August, 2, 23, 59, 59, 0, time.FixedZone("UTC+8", 8*3600))
	want := time.Date(2026, time.August, 2, 0, 0, 0, 0, time.UTC).Unix()
	if got := insuranceInventoryExitUTCDayStart(now); got != want {
		t.Fatalf("UTC day start got=%d want=%d", got, want)
	}
}
