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

type SysPermListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysPermListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysPermListLogic {
	return &SysPermListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysPermListLogic) SysPermList() (resp *types.SysPermListResp, err error) {
	result, err := l.svcCtx.SystemCli.SysPermList(l.ctx, &system.Empty{})
	if err != nil {
		return nil, err
	}
	data := make([]types.SysPermItem, 0)
	for _, item := range result.Data {
		data = append(data, types.SysPermItem{
			Key:  item.PermKey,
			Name: item.Name,
		})
	}
	return &types.SysPermListResp{
		Code: result.Code,
		Msg:  result.Msg,
		Data: data,
	}, nil
}
