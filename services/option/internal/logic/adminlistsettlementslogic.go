package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListSettlementsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminListSettlementsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListSettlementsLogic {
	return &AdminListSettlementsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询到期结算记录列表
func (l *AdminListSettlementsLogic) AdminListSettlements(in *option.ListSettlementsReq) (*option.ListSettlementsResp, error) {
	// todo: add your logic here and delete this line

	return &option.ListSettlementsResp{}, nil
}
