package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserBanksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUserBanksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserBanksLogic {
	return &ListUserBanksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 银行卡相关接口
func (l *ListUserBanksLogic) ListUserBanks(in *user.ListUserBanksReq) (*user.ListUserBanksResp, error) {
	// todo: add your logic here and delete this line

	return &user.ListUserBanksResp{}, nil
}
