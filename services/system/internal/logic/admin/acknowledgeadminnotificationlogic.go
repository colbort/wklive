package adminlogic

import (
	"context"
	"strings"

	"wklive/common/helper"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AcknowledgeAdminNotificationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAcknowledgeAdminNotificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AcknowledgeAdminNotificationLogic {
	return &AcknowledgeAdminNotificationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AcknowledgeAdminNotificationLogic) AcknowledgeAdminNotification(in *system.AcknowledgeAdminNotificationReq) (*system.AdminNotificationIncidentResp, error) {
	if in == nil ||
		strings.TrimSpace(in.EventType) == "" ||
		strings.TrimSpace(in.AlertKey) == "" ||
		in.UserId <= 0 ||
		strings.TrimSpace(in.Username) == "" ||
		strings.TrimSpace(in.Reason) == "" ||
		in.AcknowledgedAt <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid admin notification acknowledgement")
	}
	if len(in.EventType) > 64 || len(in.AlertKey) > 191 ||
		len(in.Username) > 64 || len(in.Reason) > 255 {
		return nil, status.Error(codes.InvalidArgument, "admin notification acknowledgement exceeds storage limits")
	}
	value, err := l.svcCtx.AdminNotificationModel.Acknowledge(
		l.ctx,
		models.AcknowledgeAdminNotificationIncidentInput{
			TenantId:  in.TenantId,
			EventType: strings.TrimSpace(in.EventType),
			AlertKey:  strings.TrimSpace(in.AlertKey),
			UserId:    in.UserId,
			Username:  strings.TrimSpace(in.Username),
			Reason:    strings.TrimSpace(in.Reason),
			Now:       in.AcknowledgedAt,
		},
	)
	if err != nil {
		return nil, err
	}
	return &system.AdminNotificationIncidentResp{
		Base: helper.OkResp(),
		Data: adminNotificationIncidentToProto(value),
	}, nil
}
