package internallogic

import (
	"context"
	"database/sql"
	"errors"

	"wklive/common/utils"
	"wklive/proto/common"
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
	resp := &trade.CheckOrderRiskResp{Passed: 1}
	rejectCode := ""
	rejectMsg := ""
	checkResult := trade.RiskCheckResult_RISK_CHECK_RESULT_PASS
	symbol, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, in.SymbolId)
	if err != nil {
		return nil, err
	}
	limitCfg, err := l.svcCtx.RiskUserTradeLimitModel.FindOneByTenantIdUserIdProductType(l.ctx, in.TenantId, in.UserId, symbol.ProductType)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if limitCfg != nil {
		if limitCfg.TradeEnabled == int64(common.Enable_ENABLE_DISABLED) {
			resp.Passed = 0
			rejectCode = "TRADE_DISABLED"
			rejectMsg = "trade disabled"
		} else if in.Side == common.Side_SIDE_BUY && limitCfg.CanOpen == 0 {
			resp.Passed = 0
			rejectCode = "OPEN_DISABLED"
			rejectMsg = "open disabled"
		}
	}
	symbolLimit, err := l.svcCtx.RiskUserSymbolLimitModel.FindOneByTenantIdUserIdSymbolId(l.ctx, in.TenantId, in.UserId, in.SymbolId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	qty := mustParseFloat(in.Qty)
	amount := mustParseFloat(in.Amount)
	if symbolLimit != nil && resp.Passed == 1 {
		if symbolLimit.MinOrderQty.IsPositive() && qty.IsPositive() && qty.LessThan(symbolLimit.MinOrderQty) {
			resp.Passed = 0
			rejectCode = "MIN_QTY"
			rejectMsg = "quantity below minimum"
		}
		if symbolLimit.MaxOrderQty.IsPositive() && qty.IsPositive() && qty.GreaterThan(symbolLimit.MaxOrderQty) {
			resp.Passed = 0
			rejectCode = "MAX_QTY"
			rejectMsg = "quantity exceeds maximum"
		}
		if symbolLimit.MinOrderNotional.IsPositive() && amount.IsPositive() && amount.LessThan(symbolLimit.MinOrderNotional) {
			resp.Passed = 0
			rejectCode = "MIN_NOTIONAL"
			rejectMsg = "amount below minimum"
		}
		if symbolLimit.MaxOrderNotional.IsPositive() && amount.IsPositive() && amount.GreaterThan(symbolLimit.MaxOrderNotional) {
			resp.Passed = 0
			rejectCode = "MAX_NOTIONAL"
			rejectMsg = "amount exceeds maximum"
		}
	}
	if resp.Passed == 0 {
		checkResult = trade.RiskCheckResult_RISK_CHECK_RESULT_REJECT
	}
	resp.RejectCode = rejectCode
	resp.RejectMsg = rejectMsg
	_, err = l.svcCtx.RiskOrderCheckLogModel.Insert(l.ctx, &models.TRiskOrderCheckLog{
		TenantId:      in.TenantId,
		UserId:        in.UserId,
		SymbolId:      in.SymbolId,
		ProductType:   symbol.ProductType,
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
