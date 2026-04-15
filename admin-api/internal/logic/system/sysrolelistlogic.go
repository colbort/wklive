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
	result, err := l.svcCtx.SystemCli.SysRoleList(l.ctx, &system.SysRoleListReq{
		Page: &common.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
		Status: toCommonStatus(req.Status),
	})
	if err != nil {
		return nil, err
	}

	var data []types.SysRoleItem
	for _, item := range result.Data {
		data = append(data, types.SysRoleItem{
			Id:          item.Id,
			Name:        item.Name,
			Code:        item.Code,
			Status:      fromCommonStatus(item.Status),
			Remark:      item.Remark,
			CreateTimes: item.CreateTimes,
		})
	}

	resp = &types.SysRoleListResp{
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
