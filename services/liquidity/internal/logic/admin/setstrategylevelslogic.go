package adminlogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetStrategyLevelsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetStrategyLevelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetStrategyLevelsLogic {
	return &SetStrategyLevelsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetStrategyLevelsLogic) SetStrategyLevels(in *liquidity.SetStrategyLevelsReq) (*liquidity.CommonResp, error) {
	// todo: add your logic here and delete this line

	return &liquidity.CommonResp{}, nil
}
