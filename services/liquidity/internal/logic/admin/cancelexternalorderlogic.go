package adminlogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelExternalOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelExternalOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelExternalOrderLogic {
	return &CancelExternalOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CancelExternalOrderLogic) CancelExternalOrder(in *liquidity.CancelExternalOrderReq) (*liquidity.CommonResp, error) {
	// todo: add your logic here and delete this line

	return &liquidity.CommonResp{}, nil
}
