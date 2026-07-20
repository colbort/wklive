package logic

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"
)

type ReverseInsuranceCoverLogic struct {
	ctx context.Context
	svc *svc.ServiceContext
}

func NewReverseInsuranceCoverLogic(c context.Context, s *svc.ServiceContext) *ReverseInsuranceCoverLogic {
	return &ReverseInsuranceCoverLogic{c, s}
}
func (l *ReverseInsuranceCoverLogic) Reverse(in *asset.ReverseInsuranceCoverReq) (*asset.ChangeAssetResp, error) {
	if in.TenantId <= 0 || in.LiquidationNo == "" || in.ReversalNo == "" {
		return nil, fmt.Errorf("invalid insurance reversal request")
	}
	now := utils.NowMillis()
	var after *models.TUserAsset
	err := l.svc.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		cm := models.NewTAssetInsuranceCoverModel(conn)
		am := models.NewTUserAssetModel(conn, l.svc.Config.CacheRedis)
		fm := models.NewTAssetFlowModel(conn, l.svc.Config.CacheRedis)
		im := models.NewTAssetIdempotentModel(conn, l.svc.Config.CacheRedis)
		cover, err := cm.FindOneForUpdate(ctx, in.TenantId, in.LiquidationNo)
		if err != nil {
			return err
		}
		if cover.Status == 2 {
			after, err = am.FindOneByTenantIdUserIdWalletTypeCoin(ctx, cover.TenantId, cover.FundUserId, cover.WalletType, cover.Coin)
			return err
		}
		done, err := prepareAssetIdempotent(ctx, im, in.TenantId, assetBizType(asset.BizType_BIZ_TYPE_INSURANCE_FUND), assetSceneType(asset.SceneType_SCENE_TYPE_INSURANCE_FUND_REVERSAL), in.ReversalNo, in.Remark, now)
		if err != nil {
			return err
		}
		if done {
			after, err = am.FindOneByTenantIdUserIdWalletTypeCoin(ctx, cover.TenantId, cover.FundUserId, cover.WalletType, cover.Coin)
			return err
		}
		before, err := am.FindOneForUpdate(ctx, cover.TenantId, cover.FundUserId, cover.WalletType, cover.Coin)
		if err != nil {
			return err
		}
		if cover.CoveredAmount.IsPositive() {
			if _, err = am.AddAvailableAmount(ctx, cover.TenantId, cover.FundUserId, cover.WalletType, cover.Coin, cover.CoveredAmount, now); err != nil {
				return err
			}
		}
		after, err = am.FindOneByTenantIdUserIdWalletTypeCoin(ctx, cover.TenantId, cover.FundUserId, cover.WalletType, cover.Coin)
		if err != nil {
			return err
		}
		if cover.CoveredAmount.IsPositive() {
			flow := buildAssetFlowRecord(l.svc, ctx, cover.TenantId, cover.FundUserId, cover.WalletType, cover.Coin, assetSceneType(asset.SceneType_SCENE_TYPE_INSURANCE_FUND_REVERSAL), assetBizType(asset.BizType_BIZ_TYPE_INSURANCE_FUND), assetSceneType(asset.SceneType_SCENE_TYPE_INSURANCE_FUND_REVERSAL), cover.LiquidationId, in.ReversalNo, asset.AssetOpType_ASSET_OP_TYPE_ADD, cover.CoveredAmount, before, after, in.Remark, now)
			if _, err = fm.Insert(ctx, flow); err != nil {
				return err
			}
		}
		if err = cm.MarkReversed(ctx, cover.Id, now); err != nil {
			return err
		}
		return completeAssetIdempotent(ctx, im, in.TenantId, assetBizType(asset.BizType_BIZ_TYPE_INSURANCE_FUND), assetSceneType(asset.SceneType_SCENE_TYPE_INSURANCE_FUND_REVERSAL), in.ReversalNo, now)
	})
	if err != nil {
		return nil, err
	}
	return &asset.ChangeAssetResp{Base: helper.OkResp(), Data: &asset.ChangeAssetData{BizNo: in.ReversalNo, Asset: toUserAssetProto(after)}}, nil
}
