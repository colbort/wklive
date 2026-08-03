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
	"wklive/services/asset/internal/logic/helpers"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CollectLockedRevenueLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCollectLockedRevenueLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CollectLockedRevenueLogic {
	return &CollectLockedRevenueLogic{ctx: ctx, svcCtx: svcCtx}
}

// CollectLockedRevenue atomically consumes a user's locked principal and
// credits the tenant fee revenue account.
func (l *CollectLockedRevenueLogic) CollectLockedRevenue(in *asset.CollectLockedRevenueReq) (*asset.PlatformTransferResp, error) {
	amount, err := conv.ParseDecimalField(in.GetAmount())
	accountType := strings.ToUpper(strings.TrimSpace(in.GetPlatformAccountType()))
	bizType, sceneType, bizNo := helpers.AssetBizType(in.GetBizType()), helpers.AssetSceneType(in.GetSceneType()), strings.TrimSpace(in.GetBizNo())
	targetBizType, targetBizNo := helpers.AssetBizType(in.GetTargetBizType()), strings.TrimSpace(in.GetTargetBizNo())
	if err != nil || !amount.IsPositive() || in.GetTenantId() <= 0 || accountType != feeRevenueAccountType ||
		bizType == "" || sceneType == "" || bizNo == "" || targetBizType == "" || targetBizNo == "" {
		return nil, fmt.Errorf("invalid locked revenue request")
	}

	var after *models.TUserAsset
	var accountID int64
	var platformBalance decimal.Decimal
	replay := false
	now := utils.NowMillis()
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		accounts := models.NewTAssetPlatformAccountModel(conn, l.svcCtx.Config.CacheRedis)
		platformFlows := models.NewTAssetPlatformFlowModel(conn, l.svcCtx.Config.CacheRedis)
		userAssets := models.NewTUserAssetModel(conn, l.svcCtx.Config.CacheRedis)
		locks := models.NewTAssetLockModel(conn, l.svcCtx.Config.CacheRedis)
		assetFlows := models.NewTAssetFlowModel(conn, l.svcCtx.Config.CacheRedis)
		idempotents := models.NewTAssetIdempotentModel(conn, l.svcCtx.Config.CacheRedis)

		lock, e := locks.FindOneByTenantBizNo(ctx, in.GetTenantId(), targetBizType, targetBizNo)
		if e != nil {
			return e
		}
		account, e := accounts.FindOneForUpdate(ctx, in.GetTenantId(), accountType, lock.Coin)
		if e != nil {
			return fmt.Errorf("fee revenue platform account is not configured: %w", e)
		}
		accountID, platformBalance = account.Id, account.AvailableAmount
		done, e := helpers.PrepareAssetIdempotent(ctx, idempotents, in.GetTenantId(), bizType, sceneType, bizNo, in.GetRemark(), now)
		if e != nil {
			return e
		}
		if done {
			flow, findErr := platformFlows.FindOneByTenantIdPlatformAccountIdSceneTypeBizNo(ctx, in.GetTenantId(), account.Id, sceneType, bizNo)
			if findErr != nil || flow.OpType != 1 || !flow.Amount.Equal(amount) || flow.BizType != bizType || flow.BizId != in.GetBizId() {
				return fmt.Errorf("locked revenue idempotency parameters changed")
			}
			after, e = userAssets.FindOneByTenantIdUserIdWalletTypeCoin(ctx, lock.TenantId, lock.UserId, lock.WalletType, lock.Coin)
			platformBalance, replay = flow.AfterAvailable, true
			return e
		}

		before, e := userAssets.FindOneByTenantIdUserIdWalletTypeCoin(ctx, lock.TenantId, lock.UserId, lock.WalletType, lock.Coin)
		if e != nil {
			return e
		}
		if ok, changeErr := userAssets.DeductLockedAmount(ctx, lock.TenantId, lock.UserId, lock.WalletType, lock.Coin, amount, now); changeErr != nil {
			return changeErr
		} else if !ok {
			return i18n.StatusError(ctx, i18n.DeductLockedFailed)
		}
		if ok, changeErr := locks.UpdateDeduct(ctx, lock.LockNo, amount, now); changeErr != nil {
			return changeErr
		} else if !ok {
			return i18n.StatusError(ctx, i18n.LockRecordUpdateFailed)
		}
		if e = accounts.AddAvailable(ctx, account.Id, amount, now); e != nil {
			return e
		}
		platformBalance = account.AvailableAmount.Add(amount)
		after, e = userAssets.FindOneByTenantIdUserIdWalletTypeCoin(ctx, lock.TenantId, lock.UserId, lock.WalletType, lock.Coin)
		if e != nil {
			return e
		}
		userFlow := helpers.BuildAssetFlowRecord(l.svcCtx, ctx, lock.TenantId, lock.UserId, lock.WalletType, lock.Coin, sceneType, bizType, sceneType, in.GetBizId(), bizNo, asset.AssetOpType_ASSET_OP_TYPE_LOCK_DEDUCT, amount, before, after, in.GetRemark(), now)
		if userFlow == nil {
			return fmt.Errorf("generate asset flow number failed")
		}
		if _, e = assetFlows.Insert(ctx, userFlow); e != nil {
			return e
		}
		if _, e = platformFlows.Insert(ctx, &models.TAssetPlatformFlow{
			TenantId: in.GetTenantId(), PlatformAccountId: account.Id, AccountType: accountType, Coin: lock.Coin,
			OpType: 1, Amount: amount, BeforeAvailable: account.AvailableAmount, AfterAvailable: platformBalance,
			BizType: bizType, SceneType: sceneType, BizId: in.GetBizId(), BizNo: bizNo, Remark: in.GetRemark(), CreateTimes: now,
		}); e != nil {
			return e
		}
		return helpers.CompleteAssetIdempotent(ctx, idempotents, in.GetTenantId(), bizType, sceneType, bizNo, now)
	})
	if err != nil {
		return nil, err
	}
	return &asset.PlatformTransferResp{Base: helper.OkResp(), Asset: helpers.ToUserAssetProto(after), PlatformAccountId: accountID, PlatformAccountBalance: platformBalance.String(), IdempotentReplay: replay}, nil
}
