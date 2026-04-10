package logic

import (
	"context"

	"wklive/common/pageutil"
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

	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}

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
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastID),
		Data: data,
	}, nil
}
