package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/asset"
	"wklive/services/asset/internal/helpers"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PageAssetLocksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPageAssetLocksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageAssetLocksLogic {
	return &PageAssetLocksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询锁仓明细
func (l *PageAssetLocksLogic) PageAssetLocks(in *asset.PageAssetLocksReq) (*asset.PageAssetLocksResp, error) {
	locks, total, err := l.svcCtx.AssetLockModel.FindPageByFilter(l.ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, helpers.AssetBizType(in.BizType), in.BizNo, int64(in.Status), in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}

	resp := &asset.PageAssetLocksResp{Base: helper.OkResp()}
	resp.Base.Total = total
	if int64(len(locks)) == in.Page.Limit && in.Page.Limit > 0 {
		resp.Base.HasNext = true
		resp.Base.NextCursor = locks[len(locks)-1].Id
	}

	for _, item := range locks {
		resp.Data = append(resp.Data, helpers.ToAssetLockProto(item))
	}
	return resp, nil
}
