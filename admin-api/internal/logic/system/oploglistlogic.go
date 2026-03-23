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
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
		Username: req.Username,
		Method:   req.Method,
		Path:     req.Path,
	})
	if err != nil {
		return nil, err
	}
	data := make([]types.OpLogItem, 0)
	for _, item := range result.Data {
		data = append(data, types.OpLogItem{
			Id:        item.Id,
			UserId:    item.UserId,
			Username:  item.Username,
			Method:    item.Method,
			Path:      item.Path,
			Req:       item.Req,
			Resp:      item.Resp,
			Ip:        item.Ip,
			CostMs:    item.CostMs,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
	}
	return &types.OpLogListResp{
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
