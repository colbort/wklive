package tradeadminlogic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetInsuranceFundAccountListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetInsuranceFundAccountListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInsuranceFundAccountListLogic {
	return &GetInsuranceFundAccountListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetInsuranceFundAccountListLogic) GetInsuranceFundAccountList(in *trade.GetInsuranceFundAccountListReq) (*trade.GetInsuranceFundAccountListResp, error) {
	return NewAdminInsuranceSnapshotLogic(l.ctx, l.svcCtx).GetInsuranceFundAccountList(in)
}
