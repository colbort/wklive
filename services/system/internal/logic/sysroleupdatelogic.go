package logic

import (
	"context"
	"time"

	"wklive/rpc/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysRoleUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysRoleUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysRoleUpdateLogic {
	return &SysRoleUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SysRoleUpdateLogic) SysRoleUpdate(in *system.SysRoleUpdateReq) (*system.SimpleResp, error) {
	err := l.svcCtx.RoleModel.Update(l.ctx, &models.SysRole{
		Id:        in.Id,
		Name:      in.Name,
		Status:    in.Status,
		Remark:    in.Remark,
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return nil, err
	}
	return &system.SimpleResp{
		Code: 200,
		Msg:  "更新成功",
	}, nil
}
