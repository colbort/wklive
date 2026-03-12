package logic

import (
	"context"

	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type OpLogListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOpLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OpLogListLogic {
	return &OpLogListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *OpLogListLogic) OpLogList(in *system.OpLogListReq) (*system.OpLogListResp, error) {
	page := in.Page.Page
	if page <= 0 {
		page = 1
	}
	pageSize := in.Page.Size
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	items, total, err := l.svcCtx.OpLogModel.FindPage(
		l.ctx,
		in.Username,
		in.Method,
		in.Path,
		page,
		pageSize,
	)
	if err != nil {
		return nil, err
	}

	data := make([]*system.OpLogItem, 0, len(items))
	for _, item := range items {
		data = append(data, &system.OpLogItem{
			Id:        item.Id,
			UserId:    item.UserId.Int64,
			Username:  item.Username.String,
			Method:    item.Method.String,
			Path:      item.Path.String,
			Req:       item.Req.String,
			Resp:      item.Resp.String,
			Ip:        item.Ip.String,
			CostMs:    item.CostMs.Int64,
			CreatedAt: item.CreatedAt.UnixMilli(),
			UpdatedAt: item.UpdatedAt.UnixMilli(),
		})
	}

	return &system.OpLogListResp{
		Base: &system.RespBase{
			Code:  200,
			Msg:   "success",
			Total: total,
		},
		Data: data,
	}, nil
}
