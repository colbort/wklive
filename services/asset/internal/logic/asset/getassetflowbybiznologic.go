package assetlogic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/asset"
	"wklive/services/asset/internal/logic/helpers"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAssetFlowByBizNoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAssetFlowByBizNoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAssetFlowByBizNoLogic {
	return &GetAssetFlowByBizNoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 按幂等业务键查询资产流水，用于业务服务对账。
func (l *GetAssetFlowByBizNoLogic) GetAssetFlowByBizNo(in *asset.GetAssetFlowByBizNoReq) (*asset.GetAssetFlowByBizNoResp, error) {
	if in.TenantId <= 0 || in.BizNo == "" ||
		in.BizType == asset.BizType_BIZ_TYPE_UNKNOWN ||
		in.SceneType == asset.SceneType_SCENE_TYPE_UNKNOWN {
		return &asset.GetAssetFlowByBizNoResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	item, err := l.svcCtx.AssetFlowModel.FindOneByTenantBizNo(l.ctx, in.TenantId, in.BizNo)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &asset.GetAssetFlowByBizNoResp{Base: helper.ErrResp(i18n.BusinessDataNotFound, i18n.Translate(i18n.BusinessDataNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if item.BizType != helpers.AssetBizType(in.BizType) ||
		item.SceneType != helpers.AssetSceneType(in.SceneType) {
		return &asset.GetAssetFlowByBizNoResp{Base: helper.ErrResp(i18n.BusinessDataNotFound, i18n.Translate(i18n.BusinessDataNotFound, l.ctx))}, nil
	}
	return &asset.GetAssetFlowByBizNoResp{Base: helper.OkResp(), Data: helpers.ToAssetFlowProto(item)}, nil
}
