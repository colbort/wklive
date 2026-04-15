// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/common"
	"wklive/proto/system"

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
	result, err := l.svcCtx.SystemCli.SysUserList(l.ctx, &system.SysUserListReq{
		Page: &common.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
		Status: toCommonStatus(req.Status),
	})
	if err != nil {
		return nil, err
	}

	var data []types.SysUserItem
	for _, item := range result.Data {
		data = append(data, types.SysUserItem{
			Id:               item.Id,
			Username:         item.Username,
			Nickname:         item.Nickname,
			Status:           fromCommonStatus(item.Status),
			RoleIds:          item.RoleIds,
			CreateTimes:      item.CreateTimes,
			Google2faEnabled: item.Google2FaEnabled,
		})
	}

	resp = &types.SysUserListResp{
		RespBase: types.RespBase{
			Code:       result.Base.Code,
			Msg:        result.Base.Msg,
			Total:      result.Base.Total,
			HasNext:    result.Base.HasNext,
			HasPrev:    result.Base.HasPrev,
			NextCursor: result.Base.NextCursor,
			PrevCursor: result.Base.PrevCursor,
		},
		Data: data,
	}
	return
}
