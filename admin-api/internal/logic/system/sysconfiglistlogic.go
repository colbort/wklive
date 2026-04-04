// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/system"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysConfigListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysConfigListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysConfigListLogic {
	return &SysConfigListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysConfigListLogic) SysConfigList(req *types.SysConfigListReq) (resp *types.SysConfigListResp, err error) {
	result, err := l.svcCtx.SystemCli.SysConfigList(l.ctx, &system.SysConfigListReq{
		Page: &system.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
		Keyword: req.Keyword,
	})
	if err != nil {
		return nil, err
	}
	resp = &types.SysConfigListResp{
		RespBase: types.RespBase{
			Code:       result.Base.Code,
			Msg:        result.Base.Msg,
			Total:      result.Base.Total,
			HasNext:    result.Base.HasNext,
			HasPrev:    result.Base.HasPrev,
			NextCursor: result.Base.NextCursor,
			PrevCursor: result.Base.PrevCursor,
		},
		Data: make([]types.SysConfigItem, len(result.Data)),
	}
	for i, v := range result.Data {
		resp.Data[i] = types.SysConfigItem{
			Id:          v.Id,
			ConfigKey:   v.ConfigKey,
			ConfigValue: v.ConfigValue,
			Remark:      v.Remark,
			CreateTimes:   v.CreateTimes,
		}
	}
	return
}
