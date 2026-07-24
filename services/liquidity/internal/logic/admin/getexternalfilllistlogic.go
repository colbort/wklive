package adminlogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetExternalFillListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetExternalFillListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetExternalFillListLogic {
	return &GetExternalFillListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetExternalFillListLogic) GetExternalFillList(in *liquidity.GetExternalFillListReq) (*liquidity.GetExternalFillListResp, error) {
	return listExternalFills(l.ctx, l.svcCtx, in)
}
