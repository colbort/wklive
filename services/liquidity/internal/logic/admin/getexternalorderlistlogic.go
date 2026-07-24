package adminlogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetExternalOrderListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetExternalOrderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetExternalOrderListLogic {
	return &GetExternalOrderListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetExternalOrderListLogic) GetExternalOrderList(in *liquidity.GetExternalOrderListReq) (*liquidity.GetExternalOrderListResp, error) {
	// todo: add your logic here and delete this line

	return &liquidity.GetExternalOrderListResp{}, nil
}
