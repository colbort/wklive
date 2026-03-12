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
	page := in.Page.Page
	if page <= 0 {
		page = 1
	}
	pageSize := in.Page.Size
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	items, total, err := l.svcCtx.LoginLogModel.FindPage(
		l.ctx,
		in.Username,
		in.Success,
		page,
		pageSize,
	)
	if err != nil {
		return nil, err
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
			LoginAt:  item.LoginAt.UnixMilli(),
		})
	}

	return &system.LoginLogListResp{
		Base: &system.RespBase{
			Code:  200,
			Msg:   "success",
			Total: total,
		},
		Data: data,
	}, nil
}
