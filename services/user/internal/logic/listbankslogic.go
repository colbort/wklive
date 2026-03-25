package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListBanksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListBanksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListBanksLogic {
	return &ListBanksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 银行卡相关接口
func (l *ListBanksLogic) ListBanks(in *user.ListBanksReq) (*user.ListBanksResp, error) {
	// todo: add your logic here and delete this line

	return &user.ListBanksResp{}, nil
}
