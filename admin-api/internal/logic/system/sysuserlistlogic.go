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

type SysUserListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysUserListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysUserListLogic {
	return &SysUserListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysUserListLogic) SysUserList(req *types.SysUserListReq) (resp *types.SysUserListResp, err error) {
	response, err := l.svcCtx.SystemCli.SysUserList(l.ctx, &system.SysUserListReq{
		Page: &system.PageReq{
			Page: req.Page,
			Size: req.Size,
		},
		Status: req.Status,
	})
	if err != nil {
		return nil, err
	}

	var data []types.SysUserItem
	for _, item := range response.Data {
		data = append(data, types.SysUserItem{
			Id:               item.Id,
			Username:         item.Username,
			Nickname:         item.Nickname,
			Status:           item.Status,
			RoleIds:          item.RoleIds,
			CreatedAt:        item.CreatedAt,
			Google2faEnabled: item.Google2FaEnabled,
		})
	}

	resp = &types.SysUserListResp{
		RespBase: types.RespBase{
			Code:  response.Base.Code,
			Msg:   response.Base.Msg,
			Total: response.Base.Total,
		},
		Data: data,
	}
	return
}
