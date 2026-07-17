// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_private

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/proto/user"

	"wklive/app-api/internal/logicutil"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateGuestTransferLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateGuestTransferLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateGuestTransferLogic {
	return &CreateGuestTransferLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateGuestTransferLogic) CreateGuestTransfer(_ *types.CreateGuestTransferReq, sourceOrigin string) (resp *types.CreateGuestTransferResp, err error) {
	result, err := l.svcCtx.UserCli.CreateGuestTransfer(l.ctx, &user.CreateGuestTransferReq{SourceOrigin: sourceOrigin})
	if err != nil {
		return logicutil.SystemErrorResp[types.CreateGuestTransferResp](l.ctx, err)
	}
	data := types.CreateGuestTransferData{}
	if result.Data != nil {
		data = types.CreateGuestTransferData{Code: result.Data.Code, RedirectUrl: result.Data.RedirectUrl, ExpireAt: result.Data.ExpireAt}
	}
	return &types.CreateGuestTransferResp{RespBase: types.RespBase{Code: result.Base.Code, Msg: result.Base.Msg}, Data: data}, nil
}
