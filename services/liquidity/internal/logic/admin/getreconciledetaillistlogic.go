package adminlogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetReconcileDetailListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetReconcileDetailListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetReconcileDetailListLogic {
	return &GetReconcileDetailListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetReconcileDetailListLogic) GetReconcileDetailList(in *liquidity.GetReconcileDetailListReq) (*liquidity.GetReconcileDetailListResp, error) {
	return listReconcileDetails(l.ctx, l.svcCtx, in)
}
