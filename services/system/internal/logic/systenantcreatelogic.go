package logic

import (
	"context"
	"database/sql"
	"time"

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
	if err != nil {
		return nil, err
	}
	if tenant != nil {
		return &system.RespBase{
			Code: 400,
			Msg:  "租户编码已存在",
		}, nil
	}
	_, err = l.svcCtx.TenantMode.Insert(l.ctx, &models.SysTenant{
		TenantCode:   in.TenantCode,
		TenantName:   in.TenantName,
		Status:       in.Status,
		ExpireTime:   sql.NullTime{Time: time.UnixMilli(in.ExpireTime), Valid: true},
		ContactName:  sql.NullString{String: in.ContactName, Valid: true},
		ContactPhone: sql.NullString{String: in.ContactPhone, Valid: true},
		Remark:       sql.NullString{String: in.Remark, Valid: true},
		CreateTime:   time.Now(),
		UpdateTime:   time.Now(),
	})
	if err != nil {
		return nil, err
	}
	return &system.RespBase{
		Code: 200,
		Msg:  "创建成功",
	}, nil
}
