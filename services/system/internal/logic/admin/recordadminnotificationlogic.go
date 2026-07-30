package adminlogic

import (
	"context"
	"encoding/json"
	"strings"

	"wklive/common/helper"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RecordAdminNotificationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRecordAdminNotificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecordAdminNotificationLogic {
	return &RecordAdminNotificationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RecordAdminNotificationLogic) RecordAdminNotification(in *system.RecordAdminNotificationReq) (*system.AdminNotificationIncidentResp, error) {
	if in == nil ||
		strings.TrimSpace(in.EventType) == "" ||
		strings.TrimSpace(in.AlertKey) == "" ||
		strings.TrimSpace(in.EventId) == "" ||
		(in.State != "firing" && in.State != "resolved") ||
		strings.TrimSpace(in.Severity) == "" ||
		strings.TrimSpace(in.Title) == "" ||
		strings.TrimSpace(in.Message) == "" ||
		strings.TrimSpace(in.Source) == "" ||
		!json.Valid([]byte(in.PayloadJson)) ||
		in.EventTime <= 0 ||
		in.AckTimeoutMs < 60_000 ||
		in.AckTimeoutMs > 86_400_000 {
		return nil, status.Error(codes.InvalidArgument, "invalid admin notification event")
	}
	if len(in.EventType) > 64 || len(in.AlertKey) > 191 || len(in.EventId) > 255 ||
		len(in.Severity) > 32 || len(in.Title) > 255 || len(in.Message) > 2000 ||
		len(in.Source) > 64 {
		return nil, status.Error(codes.InvalidArgument, "admin notification event exceeds storage limits")
	}
	value, err := l.svcCtx.AdminNotificationModel.Record(
		l.ctx,
		models.RecordAdminNotificationIncidentInput{
			TenantId:     in.TenantId,
			EventType:    strings.TrimSpace(in.EventType),
			AlertKey:     strings.TrimSpace(in.AlertKey),
			EventId:      strings.TrimSpace(in.EventId),
			State:        in.State,
			Severity:     strings.TrimSpace(in.Severity),
			Title:        strings.TrimSpace(in.Title),
			Message:      strings.TrimSpace(in.Message),
			Source:       strings.TrimSpace(in.Source),
			PayloadJson:  in.PayloadJson,
			EventTime:    in.EventTime,
			AckTimeoutMs: in.AckTimeoutMs,
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
