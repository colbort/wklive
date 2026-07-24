package adminlogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHedgeTaskListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetHedgeTaskListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHedgeTaskListLogic {
	return &GetHedgeTaskListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetHedgeTaskListLogic) GetHedgeTaskList(in *liquidity.GetHedgeTaskListReq) (*liquidity.GetHedgeTaskListResp, error) {
	// todo: add your logic here and delete this line

	return &liquidity.GetHedgeTaskListResp{}, nil
}
