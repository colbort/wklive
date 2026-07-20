package logic

import (
	"context"
	"fmt"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// CoverInsuranceDeficitLogic performs an idempotent, atomic and partial debit
// from a dedicated insurance-fund asset account.
type CoverInsuranceDeficitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCoverInsuranceDeficitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CoverInsuranceDeficitLogic {
	return &CoverInsuranceDeficitLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *CoverInsuranceDeficitLogic) Cover(in *asset.CoverInsuranceDeficitReq) (*asset.CoverInsuranceDeficitResp, error) {
	requested, err := conv.ParseDecimalField(in.RequestedAmount)
	if err != nil || !requested.IsPositive() {
		return nil, i18n.StatusError(l.ctx, i18n.AmountMustBePositive)
	}
	if in.TenantId <= 0 || in.FundUserId <= 0 || in.Coin == "" || in.LiquidationId <= 0 || in.LiquidationNo == "" || in.WalletType != common.WalletType_WALLET_TYPE_CONTRACT {
		return nil, fmt.Errorf("invalid insurance fund request")
	}
	now := utils.NowMillis()
	covered, replay := decimal.Zero, false
	var after *models.TUserAsset
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		am := models.NewTUserAssetModel(conn, l.svcCtx.Config.CacheRedis)
		fm := models.NewTAssetFlowModel(conn, l.svcCtx.Config.CacheRedis)
		cm := models.NewTAssetInsuranceCoverModel(conn)
		im := models.NewTAssetIdempotentModel(conn, l.svcCtx.Config.CacheRedis)
		done, err := prepareAssetIdempotent(ctx, im, in.TenantId, assetBizType(asset.BizType_BIZ_TYPE_INSURANCE_FUND), assetSceneType(asset.SceneType_SCENE_TYPE_INSURANCE_FUND_COVER), in.LiquidationNo, in.Remark, now)
		if err != nil {
			return err
		}
		if done {
			cover, err := cm.FindOneByTenantLiquidationNo(ctx, in.TenantId, in.LiquidationNo)
			if err != nil {
				return err
			}
			if !cover.RequestedAmount.Equal(requested) || cover.FundUserId != in.FundUserId || cover.WalletType != int64(in.WalletType) || cover.Coin != in.Coin || cover.LiquidationId != in.LiquidationId {
				return fmt.Errorf("insurance idempotency parameters changed")
			}
			covered, replay = cover.CoveredAmount, true
			after, err = am.FindOneByTenantIdUserIdWalletTypeCoin(ctx, in.TenantId, in.FundUserId, int64(in.WalletType), in.Coin)
			return err
		}
		before, err := am.FindOneForUpdate(ctx, in.TenantId, in.FundUserId, int64(in.WalletType), in.Coin)
		if err != nil {
			return err
		}
		covered = insuranceCoverage(requested, before.AvailableAmount)
		if covered.IsPositive() {
			ok, err := am.SubAvailableAmount(ctx, in.TenantId, in.FundUserId, int64(in.WalletType), in.Coin, covered, now)
			if err != nil || !ok {
				if err != nil {
					return err
				}
				return fmt.Errorf("insurance fund concurrent debit rejected")
			}
		}
		after, err = am.FindOneByTenantIdUserIdWalletTypeCoin(ctx, in.TenantId, in.FundUserId, int64(in.WalletType), in.Coin)
		if err != nil {
			return err
		}
		if covered.IsPositive() {
			flow := buildAssetFlowRecord(l.svcCtx, ctx, in.TenantId, in.FundUserId, int64(in.WalletType), in.Coin, assetSceneType(asset.SceneType_SCENE_TYPE_INSURANCE_FUND_COVER), assetBizType(asset.BizType_BIZ_TYPE_INSURANCE_FUND), assetSceneType(asset.SceneType_SCENE_TYPE_INSURANCE_FUND_COVER), in.LiquidationId, in.LiquidationNo, asset.AssetOpType_ASSET_OP_TYPE_SUB, covered, before, after, in.Remark, now)
			if _, err = fm.Insert(ctx, flow); err != nil {
				return err
			}
		}
		if _, err = cm.Insert(ctx, &models.TAssetInsuranceCover{TenantId: in.TenantId, FundUserId: in.FundUserId, WalletType: int64(in.WalletType), Coin: in.Coin, LiquidationId: in.LiquidationId, LiquidationNo: in.LiquidationNo, RequestedAmount: requested, CoveredAmount: covered, RemainingAmount: requested.Sub(covered), Status: 1, CreateTimes: now, UpdateTimes: now}); err != nil {
			return err
		}
		return completeAssetIdempotent(ctx, im, in.TenantId, assetBizType(asset.BizType_BIZ_TYPE_INSURANCE_FUND), assetSceneType(asset.SceneType_SCENE_TYPE_INSURANCE_FUND_COVER), in.LiquidationNo, now)
	})
	if err != nil {
		return nil, err
	}
	return &asset.CoverInsuranceDeficitResp{Base: helper.OkResp(), RequestedAmount: requested.String(), CoveredAmount: covered.String(), RemainingAmount: requested.Sub(covered).String(), FundAsset: toUserAssetProto(after), IdempotentReplay: replay}, nil
}

func insuranceCoverage(requested, available decimal.Decimal) decimal.Decimal {
	if !requested.IsPositive() || !available.IsPositive() {
		return decimal.Zero
	}
	if available.LessThan(requested) {
		return available
	}
	return requested
}
