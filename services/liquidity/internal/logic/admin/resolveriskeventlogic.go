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

type ResolveRiskEventLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResolveRiskEventLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResolveRiskEventLogic {
	return &ResolveRiskEventLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ResolveRiskEventLogic) ResolveRiskEvent(in *liquidity.ResolveRiskEventReq) (*liquidity.CommonResp, error) {
	row, err := l.svcCtx.RiskEventModel.FindOne(l.ctx, in.RiskEventId)
	if err != nil {
		return nil, err
	}
	if in.Status != liquidity.RiskEventStatus_RISK_EVENT_STATUS_RECOVERED &&
		in.Status != liquidity.RiskEventStatus_RISK_EVENT_STATUS_CLOSED {
		return nil, fmt.Errorf("risk event can only be recovered or closed")
	}
	if row.Status == int64(liquidity.RiskEventStatus_RISK_EVENT_STATUS_CLOSED) {
		return nil, fmt.Errorf("risk event is already closed")
	}
	now := time.Now().UnixMilli()
	row.Status, row.OperatorId, row.UpdateTimes = int64(in.Status), in.OperatorId, now
	if in.Status == liquidity.RiskEventStatus_RISK_EVENT_STATUS_RECOVERED {
		row.RecoveredAt = now
	} else {
		row.ClosedAt = now
	}
	if resolution := strings.TrimSpace(in.Resolution); resolution != "" {
		row.Message = row.Message + " | resolution: " + resolution
	}
	if err := l.svcCtx.RiskEventModel.Update(l.ctx, row); err != nil {
		return nil, err
	}
	return &liquidity.CommonResp{Base: helper.OkResp()}, nil
}
