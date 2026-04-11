package logic

import (
	"context"

	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeductLockedAssetByBizNoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeductLockedAssetByBizNoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeductLockedAssetByBizNoLogic {
	return &DeductLockedAssetByBizNoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeductLockedAssetByBizNoLogic) DeductLockedAssetByBizNo(in *asset.DeductLockedAssetByBizNoReq) (*asset.ChangeAssetResp, error) {
	lock, err := findLockByBizNo(l.ctx, l.svcCtx, in.TenantId, in.TargetBizType, in.TargetBizNo)
	if err != nil {
		return nil, err
	}

	return NewDeductLockedAssetLogic(l.ctx, l.svcCtx).DeductLockedAsset(&asset.DeductLockedAssetReq{
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
