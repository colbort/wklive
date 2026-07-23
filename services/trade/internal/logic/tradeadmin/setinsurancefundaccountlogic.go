package tradeadminlogic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetInsuranceFundAccountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetInsuranceFundAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetInsuranceFundAccountLogic {
	return &SetInsuranceFundAccountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetInsuranceFundAccountLogic) SetInsuranceFundAccount(in *trade.SetInsuranceFundAccountReq) (*trade.AdminCommonResp, error) {
	return NewAdminInsuranceSnapshotLogic(l.ctx, l.svcCtx).SetInsuranceFundAccount(in)
}
