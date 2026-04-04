package logic

import (
	"context"
	"database/sql"
	"time"

	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysTenantUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysTenantUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysTenantUpdateLogic {
	return &SysTenantUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新租户
func (l *SysTenantUpdateLogic) SysTenantUpdate(in *system.SysTenantUpdateReq) (*system.RespBase, error) {
	tenant, err := l.svcCtx.TenantMode.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if tenant == nil {
		return &system.RespBase{
			Code: 400,
			Msg:  "租户不存在",
		}, nil
	}
	tenant.TenantName = in.TenantName
	tenant.Status = in.Status
	tenant.ExpireTime = in.ExpireTime
	tenant.ContactName = sql.NullString{String: in.ContactName, Valid: true}
	tenant.ContactPhone = sql.NullString{String: in.ContactPhone, Valid: true}
	tenant.Remark = sql.NullString{String: in.Remark, Valid: true}
	tenant.UpdateTimes = time.Now().UnixMilli()
	err = l.svcCtx.TenantMode.Update(l.ctx, tenant)
	if err != nil {
		return nil, err
	}

	return &system.RespBase{
		Code: 200,
		Msg:  "更新成功",
	}, nil
}
