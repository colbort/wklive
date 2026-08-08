package adminlogic

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOpLogLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateOpLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOpLogLogic {
	return &CreateOpLogLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateOpLogLogic) CreateOpLog(in *system.CreateOpLogReq) (*system.RespBase, error) {
	userID, username, err := utils.GetUserInfoFromMd(l.ctx)
	if err != nil || userID <= 0 || strings.TrimSpace(username) == "" {
		return nil, errors.New("trusted operation log actor metadata is required")
	}
	tenantID, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil || tenantID < 0 {
		return nil, errors.New("trusted operation log tenant metadata is required")
	}

	// 操作人只能由可信的 RPC metadata 覆盖。系统管理员 tenant_id=0 时，
	// 允许调用方记录本次操作实际针对的租户。
	in.UserId = userID
	in.Username = username
	if tenantID > 0 {
		in.TenantId = tenantID
	}

	now := time.Now().UnixMilli()
	_, err = l.svcCtx.OpLogModel.Insert(l.ctx, &models.SysOpLog{
		TenantId:    in.TenantId,
		UserId:      in.UserId,
		Username:    truncateOpLogValue(in.Username, 64),
		Module:      truncateOpLogValue(in.Module, 64),
		Action:      truncateOpLogValue(in.Action, 64),
		Method:      truncateOpLogValue(requestMethodToString(in.Method), 16),
		Path:        truncateOpLogValue(in.Path, 255),
		Req:         nullableOpLogText(in.Req, 8192),
		Resp:        nullableOpLogText(in.Resp, 8192),
		Ip:          truncateOpLogValue(in.Ip, 64),
		CostMs:      in.CostMs,
		CreateTimes: now,
		UpdateTimes: now,
	})
	if err != nil {
		return nil, err
	}

	return &system.RespBase{Base: helper.OkResp()}, nil
}

func nullableOpLogText(value string, limit int) sql.NullString {
	value = strings.TrimSpace(truncateOpLogValue(value, limit))
	return sql.NullString{String: value, Valid: value != ""}
}

func truncateOpLogValue(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}
