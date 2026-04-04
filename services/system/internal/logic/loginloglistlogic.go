package logic

import (
	"context"

	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogListLogic {
	return &LoginLogListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 日志
func (l *LoginLogListLogic) LoginLogList(in *system.LoginLogListReq) (*system.LoginLogListResp, error) {
	items, total, err := l.svcCtx.LoginLogModel.FindPage(
		l.ctx,
		in.Username,
		in.Success,
		in.Page.Cursor,
		in.Page.Limit,
	)
	if err != nil {
		return nil, err
	}

	prevCursor := in.Page.Cursor
	if prevCursor < 0 {
		prevCursor = 0
	}
	nextCursor := int64(0)
	if int64(len(items)) == in.Page.Limit {
		lastItem := items[len(items)-1]
		nextCursor = lastItem.Id
	}
	hasPrev := prevCursor > 0
	hasNext := int64(len(items)) == in.Page.Limit

	data := make([]*system.LoginLogItem, 0, len(items))
	for _, item := range items {
		data = append(data, &system.LoginLogItem{
			Id:       item.Id,
			UserId:   item.UserId.Int64,
			Username: item.Username.String,
			Ip:       item.Ip.String,
			Ua:       item.Ua.String,
			Success:  item.Success.Int64,
			Msg:      item.Msg.String,
			LoginAt:  item.LoginAt,
		})
	}

	return &system.LoginLogListResp{
		Base: &system.RespBase{
			Code:       200,
			Msg:        "success",
			Total:      total,
			HasNext:    hasNext,
			HasPrev:    hasPrev,
			NextCursor: nextCursor,
			PrevCursor: prevCursor,
		},
		Data: data,
	}, nil
}
