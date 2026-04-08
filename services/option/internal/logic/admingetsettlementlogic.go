package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminGetSettlementLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminGetSettlementLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetSettlementLogic {
	return &AdminGetSettlementLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个到期结算记录详情
func (l *AdminGetSettlementLogic) AdminGetSettlement(in *option.GetSettlementReq) (*option.GetSettlementResp, error) {
	// todo: add your logic here and delete this line

	return &option.GetSettlementResp{}, nil
}
