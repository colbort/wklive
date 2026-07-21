package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMarketSnapshotListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMarketSnapshotListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMarketSnapshotListLogic {
	return &GetMarketSnapshotListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetMarketSnapshotListLogic) GetMarketSnapshotList(in *trade.GetMarketSnapshotListReq) (*trade.GetMarketSnapshotListResp, error) {
	// todo: add your logic here and delete this line

	return &trade.GetMarketSnapshotListResp{}, nil
}
