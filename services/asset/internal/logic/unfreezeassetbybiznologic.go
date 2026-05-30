package logic

import (
	"context"

	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnfreezeAssetByBizNoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnfreezeAssetByBizNoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnfreezeAssetByBizNoLogic {
	return &UnfreezeAssetByBizNoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UnfreezeAssetByBizNoLogic) UnfreezeAssetByBizNo(in *asset.UnfreezeAssetByBizNoReq) (*asset.ChangeAssetResp, error) {
	freeze, err := findFreezeByBizNo(l.ctx, l.svcCtx, in.TenantId, in.TargetBizType, in.TargetBizNo)
	if err != nil {
		l.Errorf("UnfreezeAssetByBizNo find freeze failed, tenantId=%d targetBizType=%d targetBizNo=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s err=%v",
			in.TenantId, in.TargetBizType, in.TargetBizNo, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, err)
		return nil, err
	}

	return NewUnfreezeAssetLogic(l.ctx, l.svcCtx).UnfreezeAsset(&asset.UnfreezeAssetReq{
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
