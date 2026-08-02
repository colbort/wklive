package assetlogic

import (
	"context"
	"errors"
	"strings"
	"time"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/asset"
	assethelpers "wklive/services/asset/internal/logic/helpers"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const optionBackstopAccountType = "OPTION_BACKSTOP"

const (
	backstopPolicyModeDisabled    int64 = 1
	backstopPolicyModePrefunded   int64 = 2
	backstopPolicyModeCreditFloor int64 = 3
	backstopPolicyStatusApproved  int64 = 2
)

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
		return nil, status.Error(codes.InvalidArgument, "invalid option platform backstop request")
	}
	now := utils.NowMillis()
	accountID := int64(0)
	balance := requested.Neg()
	policyID := int64(0)
	policyVersion := int64(0)
	policyMode := int64(0)
	dailyUsed := decimal.Zero
	replay := false
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		accounts := models.NewTAssetPlatformAccountModel(conn, l.svcCtx.Config.CacheRedis)
		flows := models.NewTAssetPlatformFlowModel(conn, l.svcCtx.Config.CacheRedis)
		covers := models.NewTAssetBackstopCoverModel(conn, l.svcCtx.Config.CacheRedis)
		policies := models.NewTAssetBackstopPolicyModel(conn, l.svcCtx.Config.CacheRedis)
		usage := models.NewTAssetBackstopUsageDailyModel(conn, l.svcCtx.Config.CacheRedis)
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
			if i18n.IsStatusError(err, i18n.AssetRequestProcessing) {
				return backstopPrecondition("platform backstop request is already processing")
			}
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
				return backstopPrecondition("platform backstop idempotency parameters changed")
			}
			policyID, policyVersion, policyMode = cover.PolicyId, cover.PolicyVersion, cover.PolicyMode
			dailyUsed = cover.DailyUsedAfter
			accountID, balance, replay = cover.PlatformAccountId, cover.BalanceAfter, true
			return nil
		}
		account, err := accounts.FindOneForUpdate(
			ctx, in.GetTenantId(), optionBackstopAccountType, coin,
		)
		if err != nil {
			if errors.Is(err, models.ErrNotFound) {
				return backstopPrecondition("option platform backstop account is not configured")
			}
			return err
		}
		accountID = account.Id
		policy, err := policies.FindEffectiveForUpdate(ctx, in.GetTenantId(), coin, now)
		if err != nil {
			if errors.Is(err, models.ErrNotFound) {
				return backstopPrecondition("no approved effective platform backstop policy")
			}
			return err
		}
		if err = validateEffectiveBackstopPolicy(policy, requested, now); err != nil {
			return err
		}
		policyID, policyVersion, policyMode = policy.Id, policy.Version, policy.Mode
		usageDay := platformBackstopUsageDay(now)
		dailyBefore := decimal.Zero
		dailyRow, usageErr := usage.FindOneForUpdate(ctx, in.GetTenantId(), coin, usageDay)
		if usageErr == nil {
			dailyBefore = dailyRow.CoveredAmount
		} else if !errors.Is(usageErr, models.ErrNotFound) {
			return usageErr
		}
		dailyAfter := dailyBefore.Add(requested)
		if dailyAfter.GreaterThan(policy.DailyLimit) {
			return backstopPrecondition("platform backstop daily limit exceeded")
		}
		balance = account.AvailableAmount.Sub(requested)
		ok, err := accounts.SubAvailableWithFloor(ctx, account.Id, requested, policy.BalanceFloor, now)
		if err != nil || !ok {
			if err != nil {
				return err
			}
			return backstopPrecondition("option platform backstop balance floor exceeded")
		}
		if dailyRow == nil {
			result, insertErr := usage.Insert(ctx, &models.TAssetBackstopUsageDaily{
				TenantId: in.GetTenantId(), Coin: coin, UsageDay: usageDay,
				CoveredAmount: requested, LastPolicyId: policy.Id,
				CreateTimes: now, UpdateTimes: now,
			})
			if insertErr != nil {
				return insertErr
			}
			if _, insertErr = result.LastInsertId(); insertErr != nil {
				return insertErr
			}
		} else if err = usage.AddCovered(ctx, dailyRow, requested, policy.Id, now); err != nil {
			return err
		}
		dailyUsed = dailyAfter
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
			PolicyId: policy.Id, PolicyVersion: policy.Version, PolicyMode: policy.Mode,
			CoveredAmount: requested, DailyUsedBefore: dailyBefore, DailyUsedAfter: dailyAfter,
			BalanceFloor: policy.BalanceFloor, BalanceBefore: account.AvailableAmount, BalanceAfter: balance,
			Status: 1, CreateTimes: now, UpdateTimes: now,
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
		PlatformAccountBalance: balance.String(), PolicyId: policyID,
		PolicyVersion: policyVersion, PolicyMode: asset.PlatformBackstopMode(policyMode),
		DailyUsedAmount: dailyUsed.String(),
	}, nil
}

func platformBackstopUsageDay(now int64) string {
	return time.UnixMilli(now).UTC().Format("20060102")
}

func backstopPrecondition(message string) error {
	return status.Error(codes.FailedPrecondition, message)
}

func validateEffectiveBackstopPolicy(
	policy *models.TAssetBackstopPolicy,
	requested decimal.Decimal,
	now int64,
) error {
	if policy == nil || policy.Status != backstopPolicyStatusApproved ||
		policy.EffectiveFrom > now || policy.EffectiveUntil <= now {
		return backstopPrecondition("invalid or ineffective platform backstop policy")
	}
	if policy.Mode == backstopPolicyModeDisabled {
		return backstopPrecondition("platform backstop policy is disabled")
	}
	if !requested.IsPositive() || !policy.PerRequestLimit.IsPositive() ||
		!policy.DailyLimit.IsPositive() || policy.PerRequestLimit.GreaterThan(policy.DailyLimit) ||
		requested.GreaterThan(policy.PerRequestLimit) {
		return backstopPrecondition("platform backstop per-request limit exceeded")
	}
	switch policy.Mode {
	case backstopPolicyModePrefunded:
		if !policy.BalanceFloor.IsZero() {
			return backstopPrecondition("prefunded platform backstop must use a zero balance floor")
		}
	case backstopPolicyModeCreditFloor:
		if !policy.BalanceFloor.IsNegative() {
			return backstopPrecondition("credit platform backstop requires a negative balance floor")
		}
	default:
		return backstopPrecondition("unsupported platform backstop policy mode")
	}
	return nil
}
