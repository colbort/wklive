package logic

import (
	"context"

	"wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyLocksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMyLocksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyLocksLogic {
	return &ListMyLocksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询我的锁仓明细
func (l *ListMyLocksLogic) ListMyLocks(in *asset.ListMyLocksReq) (*asset.ListMyLocksResp, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	items, total, err := l.svcCtx.AssetLockModel.FindPage(l.ctx, models.AssetLockPageFilter{
		TenantId:   tenantId,
		UserId:     userId,
		WalletType: int64(in.WalletType),
		Coin:       in.Coin,
		Status:     int64(in.Status),
	}, in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}
	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}
	resp := &asset.ListMyLocksResp{Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastID)}
	for _, item := range items {
		resp.Data = append(resp.Data, toAssetLockProto(item))
	}
	return resp, nil
}
