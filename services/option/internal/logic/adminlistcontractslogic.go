package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListContractsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminListContractsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListContractsLogic {
	return &AdminListContractsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询期权合约列表
func (l *AdminListContractsLogic) AdminListContracts(in *option.ListContractsReq) (*option.ListContractsResp, error) {
	// todo: add your logic here and delete this line

	return &option.ListContractsResp{}, nil
}
