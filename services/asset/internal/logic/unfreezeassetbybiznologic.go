package logic

import (
	"context"
	"strings"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
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
	rawAmount := strings.TrimSpace(in.Amount)
	amount, err := conv.ParseFloatField(rawAmount)
	if err != nil {
		l.Errorf("UnfreezeAssetByBizNo parse amount failed, tenantId=%d targetBizType=%d targetBizNo=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s err=%v",
			in.TenantId, in.TargetBizType, in.TargetBizNo, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, err)
		return nil, err
	}
	if amount < 0 {
		err := i18n.StatusError(l.ctx, i18n.AmountMustNotBeNegative)
		l.Errorf("UnfreezeAssetByBizNo validate amount failed, tenantId=%d targetBizType=%d targetBizNo=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s err=%v",
			in.TenantId, in.TargetBizType, in.TargetBizNo, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, err)
		return nil, err
	}
	unfreezeRemaining := rawAmount == "" || amount == 0
	unfreezeAmount := rawAmount

	freeze, err := findFreezeByBizNo(l.ctx, l.svcCtx, in.TenantId, in.TargetBizType, in.TargetBizNo)
	if err != nil {
		if unfreezeRemaining {
			return &asset.ChangeAssetResp{Base: helper.OkResp(), Data: &asset.ChangeAssetData{BizNo: in.BizNo}}, nil
		}
		l.Errorf("UnfreezeAssetByBizNo find freeze failed, tenantId=%d targetBizType=%d targetBizNo=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s err=%v",
			in.TenantId, in.TargetBizType, in.TargetBizNo, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, err)
		return nil, err
	}
	if unfreezeRemaining {
		if freeze.RemainAmount <= 0 || (freeze.Status != 1 && freeze.Status != 2) {
			return &asset.ChangeAssetResp{Base: helper.OkResp(), Data: &asset.ChangeAssetData{BizNo: in.BizNo}}, nil
		}
		unfreezeAmount = conv.FloatString(freeze.RemainAmount)
	}

	return NewUnfreezeAssetLogic(l.ctx, l.svcCtx).UnfreezeAsset(&asset.UnfreezeAssetReq{
		TenantId:  in.TenantId,
		FreezeNo:  freeze.FreezeNo,
		Amount:    unfreezeAmount,
		BizType:   in.BizType,
		SceneType: in.SceneType,
		BizId:     in.BizId,
		BizNo:     in.BizNo,
		Remark:    in.Remark,
	})
}
