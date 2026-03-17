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
	out, err := l.svcCtx.SystemCli.SysConfigList(l.ctx, &system.SysConfigListReq{
		Page: &system.PageReq{
			Page: req.Page,
			Size: req.Size,
		},
		Keyword: req.Keyword,
	})
	if err != nil {
		return nil, err
	}
	resp = &types.SysConfigListResp{
		RespBase: types.RespBase{
			Code: out.Base.Code,
			Msg:  out.Base.Msg,
		},
		Data: make([]types.SysConfigItem, len(out.Data)),
	}
	for i, v := range out.Data {
		resp.Data[i] = types.SysConfigItem{
			Id:          v.Id,
			ConfigKey:   v.ConfigKey,
			ConfigValue: v.ConfigValue.String(),
			Remark:      v.Remark,
			CreatedAt:   v.CreatedAt,
		}
	}
	return
}
