package adminlogic

import (
	"context"
	"fmt"
	"time"

	"wklive/common/helper"
	"wklive/proto/common"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/logic/helpers"
	"wklive/services/liquidity/internal/svc"
	"wklive/services/liquidity/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateManualHedgeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateManualHedgeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateManualHedgeLogic {
	return &CreateManualHedgeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateManualHedgeLogic) CreateManualHedge(in *liquidity.CreateManualHedgeReq) (*liquidity.HedgeTaskResp, error) {
	if in.ConfigId <= 0 || in.ProviderId <= 0 {
		return nil, fmt.Errorf("config_id and provider_id are required")
	}
	if in.Side != common.Side_SIDE_BUY && in.Side != common.Side_SIDE_SELL {
		return nil, fmt.Errorf("invalid side")
	}
	qty, err := parseNumber("qty", in.Qty)
	if err != nil || !qty.IsPositive() {
		return nil, fmt.Errorf("qty must be positive")
	}
	targetExposure, err := parseSignedNumber("target_exposure", in.TargetExposure)
	if err != nil {
		return nil, err
	}
	config, err := l.svcCtx.SymbolConfigModel.FindOne(l.ctx, in.ConfigId)
	if err != nil {
		return nil, err
	}
	provider, err := l.svcCtx.ProviderModel.FindOne(l.ctx, in.ProviderId)
	if err != nil {
		return nil, err
	}
	if provider.ProviderType != int64(liquidity.ProviderType_PROVIDER_TYPE_EXTERNAL) {
		return nil, fmt.Errorf("external provider not found")
	}
	if provider.Status != int64(liquidity.ProviderStatus_PROVIDER_STATUS_ENABLED) {
		return nil, fmt.Errorf("external provider is disabled")
	}
	now := time.Now().UnixMilli()
	row := &models.TLiquidityHedgeTask{
		HedgeNo:  fmt.Sprintf("HDG%d", time.Now().UnixNano()),
		ConfigId: in.ConfigId, ProviderId: in.ProviderId, SymbolId: config.SymbolId,
		TriggerType:    int64(liquidity.HedgeTriggerType_HEDGE_TRIGGER_TYPE_MANUAL),
		TargetExposure: targetExposure, Side: int64(in.Side), TargetQty: qty,
		Status: int64(liquidity.HedgeStatus_HEDGE_STATUS_PENDING), Version: 1,
		CreateTimes: now, UpdateTimes: now,
	}
	result, err := l.svcCtx.HedgeTaskModel.Insert(l.ctx, row)
	if err != nil {
		return nil, err
	}
	row.Id, err = result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &liquidity.HedgeTaskResp{Base: helper.OkResp(), Data: helpers.HedgeTaskToProto(row)}, nil
}
