package adminlogic

import (
	"context"
	"fmt"
	"time"

	"wklive/common/helper"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/logic/helpers"
	"wklive/services/liquidity/internal/svc"
	"wklive/services/liquidity/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type RunReconcileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRunReconcileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RunReconcileLogic {
	return &RunReconcileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RunReconcileLogic) RunReconcile(in *liquidity.RunReconcileReq) (*liquidity.ReconcileBatchResp, error) {
	if in.ProviderId <= 0 || in.ReconcileType == liquidity.ReconcileType_RECONCILE_TYPE_UNKNOWN {
		return nil, fmt.Errorf("provider_id and reconcile_type are required")
	}
	if in.WindowStart <= 0 || in.WindowEnd <= in.WindowStart {
		return nil, fmt.Errorf("invalid reconcile window")
	}
	provider, err := l.svcCtx.ProviderModel.FindOne(l.ctx, in.ProviderId)
	if err != nil {
		return nil, err
	}
	if provider.ProviderType != int64(liquidity.ProviderType_PROVIDER_TYPE_EXTERNAL) {
		return nil, fmt.Errorf("external provider not found")
	}
	now := time.Now().UnixMilli()
	row := &models.TLiquidityReconcileBatch{
		BatchNo:    fmt.Sprintf("REC%d", time.Now().UnixNano()),
		ProviderId: in.ProviderId, ReconcileType: int64(in.ReconcileType),
		WindowStart: in.WindowStart, WindowEnd: in.WindowEnd,
		Status:      int64(liquidity.ReconcileStatus_RECONCILE_STATUS_PENDING),
		CreateTimes: now, UpdateTimes: now,
	}
	result, err := l.svcCtx.ReconcileBatchModel.Insert(l.ctx, row)
	if err != nil {
		return nil, err
	}
	row.Id, err = result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &liquidity.ReconcileBatchResp{Base: helper.OkResp(), Data: helpers.ReconcileBatchToProto(row)}, nil
}
