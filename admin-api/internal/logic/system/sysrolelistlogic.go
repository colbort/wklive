// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/rpc/system"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysRoleListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysRoleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysRoleListLogic {
	return &SysRoleListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysRoleListLogic) SysRoleList(req *types.SysRoleListReq) (resp *types.SysRoleListResp, err error) {
	response, err := l.svcCtx.SystemCli.SysRoleList(l.ctx, &system.SysRoleListReq{
		Page: &system.PageReq{
			Page: int64(req.Page),
			Size: int64(req.Size),
		},
		Status: req.Status,
	})
	if err != nil {
		return nil, err
	}

	var data []types.SysRoleItem
	for _, item := range response.Data {
		data = append(data, types.SysRoleItem{
			Id:        item.Id,
			Name:      item.Name,
			Code:      item.Code,
			Status:    item.Status,
			Remark:    item.Remark,
			CreatedAt: item.CreatedAt,
		})
	}

	resp = &types.SysRoleListResp{
		RespBase: types.RespBase{
			Code: response.Base.Code,
			Msg:  response.Base.Msg,
		},
		Data: data,
	}
	return
}
