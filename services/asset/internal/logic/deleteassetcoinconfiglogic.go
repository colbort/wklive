package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteAssetCoinConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteAssetCoinConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAssetCoinConfigLogic {
	return &DeleteAssetCoinConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除APP资产操作币种显示配置
func (l *DeleteAssetCoinConfigLogic) DeleteAssetCoinConfig(in *asset.DeleteAssetCoinConfigReq) (*asset.DeleteAssetCoinConfigResp, error) {
	old, err := l.svcCtx.AssetCoinConfigModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &asset.DeleteAssetCoinConfigResp{Base: helper.GetErrResp(i18n.AssetCoinConfigNotFound, i18n.Translate(i18n.AssetCoinConfigNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if in.TenantId != 0 && old.TenantId != in.TenantId {
		return &asset.DeleteAssetCoinConfigResp{Base: helper.GetErrResp(i18n.AssetCoinConfigNotFound, i18n.Translate(i18n.AssetCoinConfigNotFound, l.ctx))}, nil
	}

	if err := l.svcCtx.AssetCoinConfigModel.Delete(l.ctx, in.Id); err != nil {
		return nil, err
	}

	return &asset.DeleteAssetCoinConfigResp{Base: helper.OkResp()}, nil
}
