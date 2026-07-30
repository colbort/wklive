package adminlogic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ClaimDueAdminNotificationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewClaimDueAdminNotificationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClaimDueAdminNotificationsLogic {
	return &ClaimDueAdminNotificationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ClaimDueAdminNotificationsLogic) ClaimDueAdminNotifications(in *system.ClaimDueAdminNotificationsReq) (*system.ClaimDueAdminNotificationsResp, error) {
	if in == nil || in.Now <= 0 ||
		in.RepeatIntervalMs < 60_000 || in.RepeatIntervalMs > 86_400_000 ||
		in.MaxLevel <= 0 || in.MaxLevel > 10 ||
		in.Limit <= 0 || in.Limit > 100 {
		return nil, status.Error(codes.InvalidArgument, "invalid admin notification escalation claim")
	}
	values, err := l.svcCtx.AdminNotificationModel.ClaimDue(
		l.ctx,
		in.Now,
		in.RepeatIntervalMs,
		in.MaxLevel,
		in.Limit,
	)
	if err != nil {
		return nil, err
	}
	data := make([]*system.AdminNotificationIncident, 0, len(values))
	for _, value := range values {
		data = append(data, adminNotificationIncidentToProto(value))
	}
	return &system.ClaimDueAdminNotificationsResp{
		Base: helper.OkResp(),
		Data: data,
	}, nil
}
