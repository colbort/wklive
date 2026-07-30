package assetlogic

import (
	"context"
	"fmt"
	"strings"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/asset"
	assethelpers "wklive/services/asset/internal/logic/helpers"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const optionBackstopAccountType = "OPTION_BACKSTOP"

type CoverPlatformBackstopDeficitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCoverPlatformBackstopDeficitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CoverPlatformBackstopDeficitLogic {
	return &CoverPlatformBackstopDeficitLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CoverPlatformBackstopDeficitLogic) CoverPlatformBackstopDeficit(in *asset.CoverPlatformBackstopDeficitReq) (*asset.CoverPlatformBackstopDeficitResp, error) {
	requested, err := conv.ParseDecimalField(in.GetRequestedAmount())
	if err != nil || !requested.IsPositive() {
		return nil, i18n.StatusError(l.ctx, i18n.AmountMustBePositive)
	}
	coin := strings.ToUpper(strings.TrimSpace(in.GetCoin()))
	if in.GetTenantId() <= 0 || coin == "" || in.GetLiquidationId() <= 0 ||
		in.GetLiquidationNo() == "" {
		return nil, fmt.Errorf("invalid option platform backstop request")
	}
	now := utils.NowMillis()
	accountID := int64(0)
	balance := requested.Neg()
	replay := false
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		accounts := models.NewTAssetPlatformAccountModel(conn, l.svcCtx.Config.CacheRedis)
		flows := models.NewTAssetPlatformFlowModel(conn, l.svcCtx.Config.CacheRedis)
		covers := models.NewTAssetBackstopCoverModel(conn, l.svcCtx.Config.CacheRedis)
		idempotent := models.NewTAssetIdempotentModel(conn, l.svcCtx.Config.CacheRedis)
		done, err := assethelpers.PrepareAssetIdempotent(
			ctx,
			idempotent,
			in.GetTenantId(),
			assethelpers.AssetBizType(asset.BizType_BIZ_TYPE_PLATFORM_BACKSTOP),
			assethelpers.AssetSceneType(asset.SceneType_SCENE_TYPE_PLATFORM_BACKSTOP_COVER),
			in.GetLiquidationNo(),
			in.GetRemark(),
			now,
		)
		if err != nil {
			return err
		}
		if done {
			cover, err := covers.FindOneByTenantIdLiquidationNo(
				ctx, in.GetTenantId(), in.GetLiquidationNo(),
			)
			if err != nil {
				return err
			}
			if !cover.CoveredAmount.Equal(requested) || cover.Coin != coin ||
				cover.LiquidationId != in.GetLiquidationId() {
				return fmt.Errorf("platform backstop idempotency parameters changed")
			}
			account, err := accounts.FindOneForUpdate(
				ctx, in.GetTenantId(), optionBackstopAccountType, coin,
			)
			if err != nil || account.Id != cover.PlatformAccountId {
				return fmt.Errorf("platform backstop account changed")
			}
			accountID, balance, replay = account.Id, account.AvailableAmount, true
			return nil
		}
		account, err := accounts.FindOneForUpdate(
			ctx, in.GetTenantId(), optionBackstopAccountType, coin,
		)
		if err != nil {
			return fmt.Errorf("option platform backstop account is not configured: %w", err)
		}
		accountID = account.Id
		balance = account.AvailableAmount.Sub(requested)
		ok, err := accounts.SubAvailableAllowNegative(ctx, account.Id, requested, now)
		if err != nil || !ok {
			if err != nil {
				return err
			}
			return fmt.Errorf("option platform backstop debit rejected")
		}
		if _, err := flows.Insert(ctx, &models.TAssetPlatformFlow{
			TenantId: in.GetTenantId(), PlatformAccountId: account.Id,
			AccountType: account.AccountType, Coin: coin, OpType: 2,
			Amount: requested, BeforeAvailable: account.AvailableAmount, AfterAvailable: balance,
			BizType:   assethelpers.AssetBizType(asset.BizType_BIZ_TYPE_PLATFORM_BACKSTOP),
			SceneType: assethelpers.AssetSceneType(asset.SceneType_SCENE_TYPE_PLATFORM_BACKSTOP_COVER),
			BizId:     in.GetLiquidationId(), BizNo: in.GetLiquidationNo(),
			Remark: in.GetRemark(), CreateTimes: now,
		}); err != nil {
			return err
		}
		if _, err := covers.Insert(ctx, &models.TAssetBackstopCover{
			TenantId: in.GetTenantId(), PlatformAccountId: account.Id, Coin: coin,
			LiquidationId: in.GetLiquidationId(), LiquidationNo: in.GetLiquidationNo(),
			CoveredAmount: requested, Status: 1, CreateTimes: now, UpdateTimes: now,
		}); err != nil {
			return err
		}
		return assethelpers.CompleteAssetIdempotent(
			ctx,
			idempotent,
			in.GetTenantId(),
			assethelpers.AssetBizType(asset.BizType_BIZ_TYPE_PLATFORM_BACKSTOP),
			assethelpers.AssetSceneType(asset.SceneType_SCENE_TYPE_PLATFORM_BACKSTOP_COVER),
			in.GetLiquidationNo(),
			now,
		)
	})
	if err != nil {
		return nil, err
	}
	return &asset.CoverPlatformBackstopDeficitResp{
		Base: helper.OkResp(), CoveredAmount: requested.String(),
		IdempotentReplay: replay, PlatformAccountId: accountID,
		PlatformAccountBalance: balance.String(),
	}, nil
}
