package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/asset"
	"wklive/services/asset/internal/helpers"
	"wklive/services/asset/internal/svc"

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
	locks, _, err := l.svcCtx.AssetLockModel.FindPageByFilter(l.ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, "", "", int64(in.Status), in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}
	resp := &asset.ListMyLocksResp{Base: helper.OkResp()}
	for _, item := range locks {
		resp.Data = append(resp.Data, helpers.ToAssetLockProto(item))
	}
	return resp, nil
}
