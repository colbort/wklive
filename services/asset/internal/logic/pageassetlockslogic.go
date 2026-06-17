package logic

import (
	"context"

	"wklive/common/pageutil"
	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

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
	locks, total, err := l.svcCtx.AssetLockModel.FindPage(l.ctx, models.AssetLockPageFilter{
		TenantId:   in.TenantId,
		UserId:     in.UserId,
		WalletType: int64(in.WalletType),
		Coin:       in.Coin,
		BizType:    assetBizType(in.BizType),
		BizNo:      in.BizNo,
		Status:     int64(in.Status),
	}, in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}

	lastID := int64(0)
	if len(locks) > 0 {
		lastID = locks[len(locks)-1].Id
	}

	resp := &asset.PageAssetLocksResp{Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(locks), total, lastID)}

	for _, item := range locks {
		resp.Data = append(resp.Data, toAssetLockProto(item))
	}
	return resp, nil
}
