package adminlogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResolveReconcileDifferenceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResolveReconcileDifferenceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResolveReconcileDifferenceLogic {
	return &ResolveReconcileDifferenceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ResolveReconcileDifferenceLogic) ResolveReconcileDifference(in *liquidity.ResolveReconcileDifferenceReq) (*liquidity.CommonResp, error) {
	// todo: add your logic here and delete this line

	return &liquidity.CommonResp{}, nil
}
