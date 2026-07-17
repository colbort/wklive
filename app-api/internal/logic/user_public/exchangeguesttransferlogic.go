// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_public

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/proto/user"

	"wklive/app-api/internal/logicutil"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExchangeGuestTransferLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewExchangeGuestTransferLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExchangeGuestTransferLogic {
	return &ExchangeGuestTransferLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ExchangeGuestTransferLogic) ExchangeGuestTransfer(req *types.ExchangeGuestTransferReq, currentOrigin string) (resp *types.ExchangeGuestTransferResp, err error) {
	result, err := l.svcCtx.UserCli.ExchangeGuestTransfer(l.ctx, &user.ExchangeGuestTransferReq{Code: req.Code, CurrentOrigin: currentOrigin})
	if err != nil {
		return logicutil.SystemErrorResp[types.ExchangeGuestTransferResp](l.ctx, err)
	}
	data := types.ExchangeGuestTransferData{}
	if result.Data != nil {
		data = types.ExchangeGuestTransferData{Token: result.Data.Token, DeviceId: result.Data.DeviceId, UserId: result.Data.UserId}
	}
	return &types.ExchangeGuestTransferResp{RespBase: types.RespBase{Code: result.Base.Code, Msg: result.Base.Msg}, Data: data}, nil
}
