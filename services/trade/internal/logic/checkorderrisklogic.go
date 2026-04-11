package logic

import (
	"context"
	"database/sql"
	"errors"

	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckOrderRiskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckOrderRiskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckOrderRiskLogic {
	return &CheckOrderRiskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 校验订单风控
func (l *CheckOrderRiskLogic) CheckOrderRisk(in *trade.CheckOrderRiskReq) (*trade.CheckOrderRiskResp, error) {
	resp := &trade.CheckOrderRiskResp{Passed: true}
	rejectCode := ""
	rejectMsg := ""
	checkResult := trade.RiskCheckResult_RISK_CHECK_RESULT_PASS
	limitCfg, err := l.svcCtx.RiskUserTradeLimitModel.FindOneByTenantIdUserIdMarketType(l.ctx, in.TenantId, in.UserId, int64(in.MarketType))
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if limitCfg != nil {
		if limitCfg.TradeEnabled == 0 {
			resp.Passed = false
			rejectCode = "TRADE_DISABLED"
			rejectMsg = "trade disabled"
		} else if in.Side == trade.TradeSide_TRADE_SIDE_BUY && limitCfg.CanOpen == 0 {
			resp.Passed = false
			rejectCode = "OPEN_DISABLED"
			rejectMsg = "open disabled"
		}
	}
	symbolLimit, err := l.svcCtx.RiskUserSymbolLimitModel.FindOneByTenantIdUserIdSymbolIdMarketType(l.ctx, in.TenantId, in.UserId, in.SymbolId, int64(in.MarketType))
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	qty := mustParseFloat(in.Qty)
	amount := mustParseFloat(in.Amount)
	if symbolLimit != nil && resp.Passed {
		if symbolLimit.MinOrderQty > 0 && qty > 0 && qty < symbolLimit.MinOrderQty {
			resp.Passed = false
			rejectCode = "MIN_QTY"
			rejectMsg = "quantity below minimum"
		}
		if symbolLimit.MaxOrderQty > 0 && qty > 0 && qty > symbolLimit.MaxOrderQty {
			resp.Passed = false
			rejectCode = "MAX_QTY"
			rejectMsg = "quantity exceeds maximum"
		}
		if symbolLimit.MinOrderNotional > 0 && amount > 0 && amount < symbolLimit.MinOrderNotional {
			resp.Passed = false
			rejectCode = "MIN_NOTIONAL"
			rejectMsg = "amount below minimum"
		}
		if symbolLimit.MaxOrderNotional > 0 && amount > 0 && amount > symbolLimit.MaxOrderNotional {
			resp.Passed = false
			rejectCode = "MAX_NOTIONAL"
			rejectMsg = "amount exceeds maximum"
		}
	}
	if !resp.Passed {
		checkResult = trade.RiskCheckResult_RISK_CHECK_RESULT_REJECT
	}
	resp.RejectCode = rejectCode
	resp.RejectMsg = rejectMsg
	_, err = l.svcCtx.RiskOrderCheckLogModel.Insert(l.ctx, &models.TRiskOrderCheckLog{
		TenantId:      in.TenantId,
		UserId:        in.UserId,
		SymbolId:      in.SymbolId,
		MarketType:    int64(in.MarketType),
		CheckType:     int64(trade.RiskCheckType_RISK_CHECK_TYPE_TRADE_PERMISSION),
		CheckResult:   int64(checkResult),
		RejectCode:    rejectCode,
		RejectMsg:     rejectMsg,
		RequestPrice:  mustParseFloat(in.Price),
		RequestQty:    qty,
		RequestAmount: amount,
		OperatorId:    in.UserId,
		Source:        int64(trade.SourceType_SOURCE_TYPE_USER),
		CheckSnapshot: sql.NullString{String: "", Valid: false},
		CreateTimes:   utils.NowMillis(),
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}
