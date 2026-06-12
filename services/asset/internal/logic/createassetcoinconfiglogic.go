package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateAssetCoinConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateAssetCoinConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAssetCoinConfigLogic {
	return &CreateAssetCoinConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建APP资产操作币种显示配置
func (l *CreateAssetCoinConfigLogic) CreateAssetCoinConfig(in *asset.CreateAssetCoinConfigReq) (*asset.AssetCoinConfigResp, error) {
	now := utils.NowMillis()
	data := &models.TAssetCoinConfig{
		TenantId:        in.TenantId,
		WalletType:      int64(in.WalletType),
		Coin:            in.Coin,
		Symbol:          in.Symbol,
		CoinName:        in.CoinName,
		CoinType:        assetCoinTypeValue(in.CoinType, int64(asset.AssetCoinType_ASSET_COIN_TYPE_CRYPTO)),
		ChainCode:       int64(in.ChainCode),
		IconUrl:         in.IconUrl,
		IconText:        in.IconText,
		IconBgColor:     in.IconBgColor,
		DecimalPlaces:   int64(in.DecimalPlaces),
		AppVisible:      assetCoinSwitchValue(in.AppVisible, 1),
		RechargeEnabled: assetCoinSwitchValue(in.RechargeEnabled, int64(common.Switch_SWITCH_OFF)),
		WithdrawEnabled: assetCoinSwitchValue(in.WithdrawEnabled, int64(common.Switch_SWITCH_OFF)),
		TransferEnabled: assetCoinSwitchValue(in.TransferEnabled, 1),
		Enabled:         assetCoinEnabledValue(in.Enabled, 1),
		Sort:            int64(in.Sort),
		Remark:          in.Remark,
		CreateTimes:     now,
		UpdateTimes:     now,
	}
	if data.DecimalPlaces == 0 {
		data.DecimalPlaces = 8
	}

	result, err := l.svcCtx.AssetCoinConfigModel.Insert(l.ctx, data)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err == nil {
		data.Id = id
	}

	return &asset.AssetCoinConfigResp{Base: helper.OkResp(), Data: toAssetCoinConfigProto(data)}, nil
}
