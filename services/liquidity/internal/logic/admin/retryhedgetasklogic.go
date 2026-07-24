package adminlogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RetryHedgeTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRetryHedgeTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetryHedgeTaskLogic {
	return &RetryHedgeTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RetryHedgeTaskLogic) RetryHedgeTask(in *liquidity.RetryHedgeTaskReq) (*liquidity.CommonResp, error) {
	// todo: add your logic here and delete this line

	return &liquidity.CommonResp{}, nil
}
