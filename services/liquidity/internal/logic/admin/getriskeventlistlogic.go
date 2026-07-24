package adminlogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRiskEventListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRiskEventListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRiskEventListLogic {
	return &GetRiskEventListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetRiskEventListLogic) GetRiskEventList(in *liquidity.GetRiskEventListReq) (*liquidity.GetRiskEventListResp, error) {
	// todo: add your logic here and delete this line

	return &liquidity.GetRiskEventListResp{}, nil
}
