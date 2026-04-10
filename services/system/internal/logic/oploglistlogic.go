package logic

import (
	"context"

	"wklive/common/pageutil"
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
	items, total, err := l.svcCtx.OpLogModel.FindPage(
		l.ctx,
		in.Username,
		in.Method,
		in.Path,
		in.Page.Cursor,
		in.Page.Limit,
	)
	if err != nil {
		return nil, err
	}

	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}

	data := make([]*system.OpLogItem, 0, len(items))
	for _, item := range items {
		data = append(data, &system.OpLogItem{
			Id:          item.Id,
			UserId:      item.UserId.Int64,
			Username:    item.Username.String,
			Method:      item.Method.String,
			Path:        item.Path.String,
			Req:         item.Req.String,
			Resp:        item.Resp.String,
			Ip:          item.Ip.String,
			CostMs:      item.CostMs.Int64,
			CreateTimes: item.CreateTimes,
			UpdateTimes: item.UpdateTimes,
		})
	}

	return &system.OpLogListResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastID),
		Data: data,
	}, nil
}
