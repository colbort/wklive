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

type OpLogListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOpLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OpLogListLogic {
	return &OpLogListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OpLogListLogic) OpLogList(req *types.OpLogListReq) (resp *types.OpLogListResp, err error) {
	result, err := l.svcCtx.SystemCli.OpLogList(l.ctx, &system.OpLogListReq{
		Page: &system.PageReq{
			Page: req.Page,
			Size: req.Size,
		},
		Username: req.Username,
		Module:   req.Module,
		Action:   req.Action,
	})
	if err != nil {
		return nil, err
	}
	data := make([]types.OpLogItem, 0)
	for _, item := range result.Data {
		data = append(data, types.OpLogItem{
			Id:         item.Id,
			UserId:     item.UserId,
			Username:   item.Username,
			Module:     item.Module,
			Action:     item.Action,
			Method:     item.Module,
			Path:       item.Path,
			Ip:         item.Ip,
			Ua:         item.Ua,
			RespCode:   item.RespCode,
			DurationMs: item.DurationMs,
			CreatedAt:  item.CreatedAt,
		})
	}
	return &types.OpLogListResp{
		RespBase: types.RespBase{
			Code:  result.Base.Code,
			Msg:   result.Base.Msg,
			Total: result.Base.Total,
		},
		Data: data,
	}, nil
}
