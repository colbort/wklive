package internallogic

import (
	"context"
	"fmt"
	"strings"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const insuranceFundAccountType = "INSURANCE_FUND"

type CoverInsuranceDeficitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCoverInsuranceDeficitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CoverInsuranceDeficitLogic {
	return &CoverInsuranceDeficitLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *CoverInsuranceDeficitLogic) CoverInsuranceDeficit(in *asset.CoverInsuranceDeficitReq) (*asset.CoverInsuranceDeficitResp, error) {
	requested, err := conv.ParseDecimalField(in.GetRequestedAmount())
	if err != nil || !requested.IsPositive() {
		return nil, i18n.StatusError(l.ctx, i18n.AmountMustBePositive)
	}
	coin := strings.ToUpper(strings.TrimSpace(in.GetCoin()))
	if in.GetTenantId() <= 0 || coin == "" || in.GetLiquidationId() <= 0 || in.GetLiquidationNo() == "" {
		return nil, fmt.Errorf("invalid insurance fund request")
	}
	now := utils.NowMillis()
	covered, balance := decimal.Zero, decimal.Zero
	accountID := int64(0)
	replay := false
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		accounts := models.NewTAssetPlatformAccountModel(conn, l.svcCtx.Config.CacheRedis)
		flows := models.NewTAssetPlatformFlowModel(conn, l.svcCtx.Config.CacheRedis)
		covers := models.NewTAssetInsuranceCoverModel(conn, l.svcCtx.Config.CacheRedis)
		idempotent := models.NewTAssetIdempotentModel(conn, l.svcCtx.Config.CacheRedis)
		done, err := prepareAssetIdempotent(ctx, idempotent, in.GetTenantId(), assetBizType(asset.BizType_BIZ_TYPE_INSURANCE_FUND), assetSceneType(asset.SceneType_SCENE_TYPE_INSURANCE_FUND_COVER), in.GetLiquidationNo(), in.GetRemark(), now)
		if err != nil {
			return err
		}
		if done {
			cover, err := covers.FindOneByTenantLiquidationNo(ctx, in.GetTenantId(), in.GetLiquidationNo())
			if err != nil {
				return err
			}
			if !cover.RequestedAmount.Equal(requested) || cover.Coin != coin || cover.LiquidationId != in.GetLiquidationId() {
				return fmt.Errorf("insurance idempotency parameters changed")
			}
			account, err := accounts.FindOneForUpdate(ctx, in.GetTenantId(), insuranceFundAccountType, coin)
			if err != nil || account.Id != cover.PlatformAccountId {
				return fmt.Errorf("insurance platform account changed")
			}
			accountID, covered, balance, replay = account.Id, cover.CoveredAmount, account.AvailableAmount, true
			return nil
		}
		account, err := accounts.FindOneForUpdate(ctx, in.GetTenantId(), insuranceFundAccountType, coin)
		if err != nil {
			return fmt.Errorf("insurance platform account is not configured: %w", err)
		}
		accountID = account.Id
		covered = insuranceCoverage(requested, account.AvailableAmount)
		balance = account.AvailableAmount.Sub(covered)
		if covered.IsPositive() {
			ok, err := accounts.SubAvailable(ctx, account.Id, covered, now)
			if err != nil || !ok {
				if err != nil {
					return err
				}
				return fmt.Errorf("insurance platform account concurrent debit rejected")
			}
			_, err = flows.Insert(ctx, &models.TAssetPlatformFlow{TenantId: in.GetTenantId(), PlatformAccountId: account.Id, AccountType: account.AccountType, Coin: coin, OpType: 2, Amount: covered, BeforeAvailable: account.AvailableAmount, AfterAvailable: balance, BizType: assetBizType(asset.BizType_BIZ_TYPE_INSURANCE_FUND), SceneType: assetSceneType(asset.SceneType_SCENE_TYPE_INSURANCE_FUND_COVER), BizId: in.GetLiquidationId(), BizNo: in.GetLiquidationNo(), Remark: in.GetRemark(), CreateTimes: now})
			if err != nil {
				return err
			}
		}
		_, err = covers.Insert(ctx, &models.TAssetInsuranceCover{TenantId: in.GetTenantId(), PlatformAccountId: account.Id, Coin: coin, LiquidationId: in.GetLiquidationId(), LiquidationNo: in.GetLiquidationNo(), RequestedAmount: requested, CoveredAmount: covered, RemainingAmount: requested.Sub(covered), Status: 1, CreateTimes: now, UpdateTimes: now})
		if err != nil {
			return err
		}
		return completeAssetIdempotent(ctx, idempotent, in.GetTenantId(), assetBizType(asset.BizType_BIZ_TYPE_INSURANCE_FUND), assetSceneType(asset.SceneType_SCENE_TYPE_INSURANCE_FUND_COVER), in.GetLiquidationNo(), now)
	})
	if err != nil {
		return nil, err
	}
	return &asset.CoverInsuranceDeficitResp{Base: helper.OkResp(), RequestedAmount: requested.String(), CoveredAmount: covered.String(), RemainingAmount: requested.Sub(covered).String(), IdempotentReplay: replay, PlatformAccountId: accountID, PlatformAccountBalance: balance.String()}, nil
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
