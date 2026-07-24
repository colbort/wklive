package adminlogic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"wklive/common/helper"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResolveReconcileDifferenceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResolveReconcileDifferenceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResolveReconcileDifferenceLogic {
	return &ResolveReconcileDifferenceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ResolveReconcileDifferenceLogic) ResolveReconcileDifference(in *liquidity.ResolveReconcileDifferenceReq) (*liquidity.CommonResp, error) {
	row, err := l.svcCtx.ReconcileDetailModel.FindOne(l.ctx, in.DifferenceId)
	if err != nil {
		return nil, err
	}
	if row.TenantId != in.TenantId {
		return nil, fmt.Errorf("reconcile difference not found")
	}
	switch in.Status {
	case liquidity.ReconcileDifferenceStatus_RECONCILE_DIFFERENCE_STATUS_RESOLVED,
		liquidity.ReconcileDifferenceStatus_RECONCILE_DIFFERENCE_STATUS_IGNORED,
		liquidity.ReconcileDifferenceStatus_RECONCILE_DIFFERENCE_STATUS_MANUAL_REQUIRED:
	default:
		return nil, fmt.Errorf("invalid resolution status")
	}
	if row.Status == int64(liquidity.ReconcileDifferenceStatus_RECONCILE_DIFFERENCE_STATUS_RESOLVED) ||
		row.Status == int64(liquidity.ReconcileDifferenceStatus_RECONCILE_DIFFERENCE_STATUS_IGNORED) {
		return nil, fmt.Errorf("reconcile difference is already terminal")
	}
	now := time.Now().UnixMilli()
	row.Status, row.Resolution, row.OperatorId = int64(in.Status), strings.TrimSpace(in.Resolution), in.OperatorId
	row.ResolvedAt, row.UpdateTimes = now, now
	if err := l.svcCtx.ReconcileDetailModel.Update(l.ctx, row); err != nil {
		return nil, err
	}
	return &liquidity.CommonResp{Base: helper.OkResp()}, nil
}
