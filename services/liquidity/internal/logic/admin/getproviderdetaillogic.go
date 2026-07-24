package adminlogic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/logic/helpers"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProviderDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetProviderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProviderDetailLogic {
	return &GetProviderDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetProviderDetailLogic) GetProviderDetail(in *liquidity.GetProviderDetailReq) (*liquidity.ProviderResp, error) {
	if err := helpers.RequireID("id", in.Id); err != nil {
		return nil, err
	}
	row, err := l.svcCtx.ProviderModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &liquidity.ProviderResp{Base: helper.OkResp(), Data: helpers.ProviderToProto(row)}, nil
}
