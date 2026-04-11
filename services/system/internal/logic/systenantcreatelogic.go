package logic

import (
	"context"
	"database/sql"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
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
			Base: helper.GetErrResp(400, i18n.Translate(i18n.TenantCodeAlreadyExists, l.ctx)),
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
		CreateTimes:  utils.NowMillis(),
		UpdateTimes:  utils.NowMillis(),
	})
	if err != nil {
		return nil, err
	}
	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
