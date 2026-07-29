package adminlogic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/market"
	"wklive/services/market/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChangePriceFormulaStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChangePriceFormulaStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePriceFormulaStatusLogic {
	return &ChangePriceFormulaStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ChangePriceFormulaStatusLogic) ChangePriceFormulaStatus(in *market.ChangePriceFormulaStatusReq) (*market.CommonResp, error) {
	if in == nil || in.Id <= 0 || (in.Status != 1 && in.Status != 3) {
		return nil, errors.New("status must be activate(1) or revoke(3)")
	}
	now := utils.NowMillis()
	var err error
	if in.Status == 1 {
		err = l.svcCtx.PriceFormulaModel.ActivateVersion(l.ctx, in.Id, now)
	} else {
		err = l.svcCtx.PriceFormulaModel.RevokeVersion(l.ctx, in.Id, now)
	}
	if err != nil {
		return nil, err
	}
	return &market.CommonResp{Base: helper.OkResp()}, nil
}
