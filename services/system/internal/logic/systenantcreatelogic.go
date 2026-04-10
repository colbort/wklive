package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"wklive/common/helper"
	"wklive/proto/common"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysTenantCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysTenantCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysTenantCreateLogic {
	return &SysTenantCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建租户
func (l *SysTenantCreateLogic) SysTenantCreate(in *system.SysTenantCreateReq) (*system.RespBase, error) {
	tenant, err := l.svcCtx.TenantMode.FindByTenantCode(l.ctx, in.TenantCode)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if tenant != nil {
		return &system.RespBase{
			Base: &common.RespBase{
				Code: 400,
				Msg:  "租户编码已存在",
			},
		}, nil
	}
	_, err = l.svcCtx.TenantMode.Insert(l.ctx, &models.SysTenant{
		TenantCode:   in.TenantCode,
		TenantName:   in.TenantName,
		Status:       in.Status,
		ExpireTime:   in.ExpireTime,
		ContactName:  sql.NullString{String: in.ContactName, Valid: true},
		ContactPhone: sql.NullString{String: in.ContactPhone, Valid: true},
		Remark:       sql.NullString{String: in.Remark, Valid: true},
		CreateTimes:  time.Now().UnixMilli(),
		UpdateTimes:  time.Now().UnixMilli(),
	})
	if err != nil {
		return nil, err
	}
	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
