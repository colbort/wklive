package adminlogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateManualHedgeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateManualHedgeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateManualHedgeLogic {
	return &CreateManualHedgeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateManualHedgeLogic) CreateManualHedge(in *liquidity.CreateManualHedgeReq) (*liquidity.HedgeTaskResp, error) {
	// todo: add your logic here and delete this line

	return &liquidity.HedgeTaskResp{}, nil
}
