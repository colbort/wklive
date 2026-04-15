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

type SysMenuListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysMenuListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysMenuListLogic {
	return &SysMenuListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysMenuListLogic) SysMenuList(req *types.SysMenuListReq) (resp *types.SysMenuListResp, err error) {
	result, err := l.svcCtx.SystemCli.SysMenuList(l.ctx, &system.SysMenuListReq{
		Page: &common.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
		Keyword:  req.Keyword,
		Status:   toCommonStatus(req.Status),
		Visible:  toVisibleStatus(req.Visible),
		MenuType: toMenuType(req.MenuType),
	})
	if err != nil {
		return nil, err
	}
	data := make([]types.SysMenuItem, 0)
	for _, v := range result.Data {
		data = append(data, types.SysMenuItem{
			Id:        v.Id,
			ParentId:  v.ParentId,
			Name:      v.Name,
			MenuType:  fromMenuType(v.MenuType),
			Method:    fromRequestMethod(v.Method),
			Path:      v.Path,
			Component: v.Component,
			Icon:      v.Icon,
			Sort:      v.Sort,
			Visible:   fromVisibleStatus(v.Visible),
			Status:    fromCommonStatus(v.Status),
			Perms:     v.Perms,
		})
	}
	return &types.SysMenuListResp{
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
	}, nil
}
