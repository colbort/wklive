package logic

import (
	"context"

	"wklive/common/pageutil"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMarginSnapshotListAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMarginSnapshotListAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMarginSnapshotListAdminLogic {
	return &GetMarginSnapshotListAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取合约风控保证金快照列表
func (l *GetMarginSnapshotListAdminLogic) GetMarginSnapshotListAdmin(in *trade.GetMarginSnapshotListAdminReq) (*trade.GetMarginSnapshotListAdminResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.ContractMarginSnapshotModel.FindPage(l.ctx, in.TenantId, in.UserId, in.MarginAsset, cursor, limit)
	if err != nil {
		return nil, err
	}
	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}
	resp := &trade.GetMarginSnapshotListAdminResp{Base: pageutil.Base(cursor, limit, len(items), total, lastID)}
	for _, item := range items {
		resp.Data = append(resp.Data, marginSnapshotToProto(item))
	}
	return resp, nil
}
