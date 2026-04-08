package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListAccountsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminListAccountsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListAccountsLogic {
	return &AdminListAccountsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询账户资产列表
func (l *AdminListAccountsLogic) AdminListAccounts(in *option.ListAccountsReq) (*option.ListAccountsResp, error) {
	// todo: add your logic here and delete this line

	return &option.ListAccountsResp{}, nil
}
