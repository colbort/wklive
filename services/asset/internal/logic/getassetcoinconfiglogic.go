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

type GetAssetCoinConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAssetCoinConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAssetCoinConfigLogic {
	return &GetAssetCoinConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询APP资产操作币种显示配置详情
func (l *GetAssetCoinConfigLogic) GetAssetCoinConfig(in *asset.GetAssetCoinConfigReq) (*asset.AssetCoinConfigResp, error) {
	data, err := l.svcCtx.AssetCoinConfigModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &asset.AssetCoinConfigResp{Base: helper.ErrResp(i18n.AssetCoinConfigNotFound, i18n.Translate(i18n.AssetCoinConfigNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if in.TenantId != 0 && data.TenantId != in.TenantId {
		return &asset.AssetCoinConfigResp{Base: helper.ErrResp(i18n.AssetCoinConfigNotFound, i18n.Translate(i18n.AssetCoinConfigNotFound, l.ctx))}, nil
	}

	return &asset.AssetCoinConfigResp{Base: helper.OkResp(), Data: toAssetCoinConfigProto(data)}, nil
}
