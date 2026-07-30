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

type ReleaseAdminNotificationEscalationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReleaseAdminNotificationEscalationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReleaseAdminNotificationEscalationLogic {
	return &ReleaseAdminNotificationEscalationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ReleaseAdminNotificationEscalationLogic) ReleaseAdminNotificationEscalation(in *system.ReleaseAdminNotificationEscalationReq) (*system.RespBase, error) {
	if in == nil || in.Id <= 0 || in.ClaimedLevel <= 0 ||
		in.RetryAt <= in.Now || in.Now <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid admin notification escalation release")
	}
	released, err := l.svcCtx.AdminNotificationModel.ReleaseEscalation(
		l.ctx,
		in.Id,
		in.ClaimedLevel,
		in.RetryAt,
		in.Now,
	)
	if err != nil {
		return nil, err
	}
	if !released {
		return nil, status.Error(codes.FailedPrecondition, "admin notification escalation claim is no longer active")
	}
	return &system.RespBase{Base: helper.OkResp()}, nil
}
