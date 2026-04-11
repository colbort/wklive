package logic

import (
	"context"

	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnlockAssetByBizNoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnlockAssetByBizNoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnlockAssetByBizNoLogic {
	return &UnlockAssetByBizNoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UnlockAssetByBizNoLogic) UnlockAssetByBizNo(in *asset.UnlockAssetByBizNoReq) (*asset.ChangeAssetResp, error) {
	lock, err := findLockByBizNo(l.ctx, l.svcCtx, in.TenantId, in.TargetBizType, in.TargetBizNo)
	if err != nil {
		return nil, err
	}

	return NewUnlockAssetLogic(l.ctx, l.svcCtx).UnlockAsset(&asset.UnlockAssetReq{
		TenantId:  in.TenantId,
		LockNo:    lock.LockNo,
		Amount:    in.Amount,
		BizType:   in.BizType,
		SceneType: in.SceneType,
		BizId:     in.BizId,
		BizNo:     in.BizNo,
		Remark:    in.Remark,
	})
}
