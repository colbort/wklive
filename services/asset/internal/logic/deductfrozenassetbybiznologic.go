package logic

import (
	"context"

	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeductFrozenAssetByBizNoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeductFrozenAssetByBizNoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeductFrozenAssetByBizNoLogic {
	return &DeductFrozenAssetByBizNoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeductFrozenAssetByBizNoLogic) DeductFrozenAssetByBizNo(in *asset.DeductFrozenAssetByBizNoReq) (*asset.ChangeAssetResp, error) {
	freeze, err := findFreezeByBizNo(l.ctx, l.svcCtx, in.TenantId, in.TargetBizType, in.TargetBizNo)
	if err != nil {
		return nil, err
	}

	return NewDeductFrozenAssetLogic(l.ctx, l.svcCtx).DeductFrozenAsset(&asset.DeductFrozenAssetReq{
		TenantId:  in.TenantId,
		FreezeNo:  freeze.FreezeNo,
		Amount:    in.Amount,
		BizType:   in.BizType,
		SceneType: in.SceneType,
		BizId:     in.BizId,
		BizNo:     in.BizNo,
		Remark:    in.Remark,
	})
}
