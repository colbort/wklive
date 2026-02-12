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

type LoginLogListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogListLogic {
	return &LoginLogListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogListLogic) LoginLogList(req *types.LoginLogListReq) (resp *types.LoginLogListResp, err error) {
	result, err := l.svcCtx.SystemCli.LoginLogList(l.ctx, &system.LoginLogListReq{
		Page: &system.PageReq{
			Page: req.Page,
			Size: req.Size,
		},
		Username: req.Username,
		Result:   req.Result,
	})
	if err != nil {
		return nil, err
	}
	data := make([]types.LoginLogItem, 0)
	for _, item := range result.Data {
		data = append(data, types.LoginLogItem{
			Id:        item.Id,
			UserId:    item.UserId,
			Username:  item.Username,
			Ip:        item.Ip,
			Ua:        item.Ua,
			Result:    item.Result,
			Reason:    item.Reason,
			CreatedAt: item.CreatedAt,
		})
	}
	return &types.LoginLogListResp{
		RespBase: types.RespBase{
			Code:  result.Base.Code,
			Msg:   result.Base.Msg,
			Total: result.Base.Total,
		},
		Data: data,
	}, nil
}
