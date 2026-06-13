package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
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
			return &asset.AssetCoinConfigResp{Base: helper.GetErrResp(i18n.AssetCoinConfigNotFound, i18n.Translate(i18n.AssetCoinConfigNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	allowTenantUpdate, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, old.TenantId)
	if err != nil {
		return nil, i18n.StatusError(l.ctx, i18n.UserNotFound)
	}
	if forbidden {
		return &asset.AssetCoinConfigResp{Base: helper.GetErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx))}, nil
	}
	if !allowed {
		return &asset.AssetCoinConfigResp{Base: helper.GetErrResp(i18n.AssetCoinConfigNotFound, i18n.Translate(i18n.AssetCoinConfigNotFound, l.ctx))}, nil
	}

	if allowTenantUpdate {
		old.TenantId = in.TenantId
	}
	if in.WalletType != 0 {
		old.WalletType = int64(in.WalletType)
	}
	if in.Coin != "" {
		old.Coin = in.Coin
	}
	if in.Symbol != "" {
		old.Symbol = in.Symbol
	}
	if in.CoinName != "" {
		old.CoinName = in.CoinName
	}
	if in.CoinType != asset.AssetCoinType_ASSET_COIN_TYPE_UNKNOWN {
		old.CoinType = int64(in.CoinType)
	}
	if in.ChainCode != 0 {
		old.ChainCode = int64(in.ChainCode)
	}
	if in.IconUrl != "" {
		old.IconUrl = in.IconUrl
	}
	if in.IconText != "" {
		old.IconText = in.IconText
	}
	if in.IconBgColor != "" {
		old.IconBgColor = in.IconBgColor
	}
	if in.DecimalPlaces != 0 {
		old.DecimalPlaces = int64(in.DecimalPlaces)
	}
	if in.AppVisible != common.Switch_SWITCH_UNKNOWN {
		old.AppVisible = int64(in.AppVisible)
	}
	if in.RechargeEnabled != common.Switch_SWITCH_UNKNOWN {
		old.RechargeEnabled = int64(in.RechargeEnabled)
	}
	if in.WithdrawEnabled != common.Switch_SWITCH_UNKNOWN {
		old.WithdrawEnabled = int64(in.WithdrawEnabled)
	}
	if in.TransferEnabled != common.Switch_SWITCH_UNKNOWN {
		old.TransferEnabled = int64(in.TransferEnabled)
	}
	if in.Enabled != common.Enable_ENABLE_UNKNOWN {
		old.Enabled = int64(in.Enabled)
	}
	if in.Sort != 0 {
		old.Sort = int64(in.Sort)
	}
	if in.Remark != "" {
		old.Remark = in.Remark
	}
	old.UpdateTimes = utils.NowMillis()

	if err := l.svcCtx.AssetCoinConfigModel.Update(l.ctx, old); err != nil {
		return nil, err
	}

	return &asset.AssetCoinConfigResp{Base: helper.OkResp(), Data: toAssetCoinConfigProto(old)}, nil
}
