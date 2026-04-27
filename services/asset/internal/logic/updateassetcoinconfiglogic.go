package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAssetCoinConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateAssetCoinConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAssetCoinConfigLogic {
	return &UpdateAssetCoinConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新APP资产操作币种显示配置
func (l *UpdateAssetCoinConfigLogic) UpdateAssetCoinConfig(in *asset.UpdateAssetCoinConfigReq) (*asset.AssetCoinConfigResp, error) {
	old, err := l.svcCtx.AssetCoinConfigModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &asset.AssetCoinConfigResp{Base: helper.GetErrResp(i18n.CodeNotFound, "asset coin config not found")}, nil
		}
		return nil, err
	}
	if in.TenantId != 0 && old.TenantId != in.TenantId {
		return &asset.AssetCoinConfigResp{Base: helper.GetErrResp(i18n.CodeNotFound, "asset coin config not found")}, nil
	}

	data := &models.TAssetCoinConfig{
		Id:              old.Id,
		TenantId:        in.TenantId,
		WalletType:      int64(in.WalletType),
		Coin:            in.Coin,
		Symbol:          in.Symbol,
		CoinName:        in.CoinName,
		CoinType:        assetCoinTypeValue(in.CoinType, old.CoinType),
		IconUrl:         in.IconUrl,
		IconText:        in.IconText,
		IconBgColor:     in.IconBgColor,
		DecimalPlaces:   int64(in.DecimalPlaces),
		AppVisible:      assetCoinSwitchValue(in.AppVisible, old.AppVisible),
		RechargeEnabled: assetCoinSwitchValue(in.RechargeEnabled, old.RechargeEnabled),
		WithdrawEnabled: assetCoinSwitchValue(in.WithdrawEnabled, old.WithdrawEnabled),
		TransferEnabled: assetCoinSwitchValue(in.TransferEnabled, old.TransferEnabled),
		Status:          assetCoinStatusValue(in.Status, old.Status),
		Sort:            int64(in.Sort),
		Remark:          in.Remark,
		CreateTimes:     old.CreateTimes,
		UpdateTimes:     utils.NowMillis(),
	}
	if data.TenantId == 0 {
		data.TenantId = old.TenantId
	}
	if data.WalletType == 0 {
		data.WalletType = old.WalletType
	}
	if data.Coin == "" {
		data.Coin = old.Coin
	}
	if data.Symbol == "" {
		data.Symbol = old.Symbol
	}
	if data.CoinName == "" {
		data.CoinName = old.CoinName
	}
	if data.DecimalPlaces == 0 {
		data.DecimalPlaces = old.DecimalPlaces
	}

	if err := l.svcCtx.AssetCoinConfigModel.Update(l.ctx, data); err != nil {
		return nil, err
	}

	return &asset.AssetCoinConfigResp{Base: helper.OkResp(), Data: toAssetCoinConfigProto(data)}, nil
}
