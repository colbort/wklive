package tradeapplogic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMarginSnapshotListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMarginSnapshotListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMarginSnapshotListLogic {
	return &GetMarginSnapshotListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取合约风控保证金快照列表
func (l *GetMarginSnapshotListLogic) GetMarginSnapshotList(in *trade.GetMarginSnapshotListReq) (*trade.GetMarginSnapshotListResp, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	items, _, err := l.svcCtx.ContractMarginSnapshotModel.FindPage(l.ctx, tenantId, userId, in.MarginAsset, 0, 1000)
	if err != nil {
		return nil, err
	}
	resp := &trade.GetMarginSnapshotListResp{Base: helper.OkResp()}
	for _, item := range items {
		resp.Data = append(resp.Data, marginSnapshotToProto(item))
	}
	return resp, nil
}
