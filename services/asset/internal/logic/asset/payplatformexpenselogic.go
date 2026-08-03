package assetlogic

import (
	"context"
	"fmt"
	"strings"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/services/asset/internal/logic/helpers"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const stakingRewardAccountType = "STAKING_REWARD"

type PayPlatformExpenseLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPayPlatformExpenseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PayPlatformExpenseLogic {
	return &PayPlatformExpenseLogic{ctx: ctx, svcCtx: svcCtx}
}

// PayPlatformExpense moves value from a tenant-owned funding account to a
// user wallet in one local transaction. Both sides share the same idempotency
// fence, so a retry can neither mint twice nor debit twice.
func (l *PayPlatformExpenseLogic) PayPlatformExpense(in *asset.PayPlatformExpenseReq) (*asset.PlatformTransferResp, error) {
	amount, err := conv.ParseDecimalField(in.GetAmount())
	accountType := strings.ToUpper(strings.TrimSpace(in.GetPlatformAccountType()))
	coin := strings.ToUpper(strings.TrimSpace(in.GetCoin()))
	bizType, sceneType, bizNo := helpers.AssetBizType(in.GetBizType()), helpers.AssetSceneType(in.GetSceneType()), strings.TrimSpace(in.GetBizNo())
	if err != nil || !amount.IsPositive() || in.GetTenantId() <= 0 || in.GetUserId() <= 0 ||
		in.GetWalletType() == 0 || accountType != stakingRewardAccountType || coin == "" || bizType == "" || sceneType == "" || bizNo == "" {
		return nil, fmt.Errorf("invalid platform expense request")
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
		assetFlows := models.NewTAssetFlowModel(conn, l.svcCtx.Config.CacheRedis)
		idempotents := models.NewTAssetIdempotentModel(conn, l.svcCtx.Config.CacheRedis)

		account, e := accounts.FindOneForUpdate(ctx, in.GetTenantId(), accountType, coin)
		if e != nil {
			return fmt.Errorf("staking reward platform account is not configured: %w", e)
		}
		accountID, platformBalance = account.Id, account.AvailableAmount
		done, e := helpers.PrepareAssetIdempotent(ctx, idempotents, in.GetTenantId(), bizType, sceneType, bizNo, in.GetRemark(), now)
		if e != nil {
			return e
		}
		if done {
			flow, findErr := platformFlows.FindOneByTenantIdPlatformAccountIdSceneTypeBizNo(ctx, in.GetTenantId(), account.Id, sceneType, bizNo)
			if findErr != nil || flow.OpType != 2 || !flow.Amount.Equal(amount) || flow.BizType != bizType || flow.BizId != in.GetBizId() {
				return fmt.Errorf("platform expense idempotency parameters changed")
			}
			after, e = userAssets.FindOneByTenantIdUserIdWalletTypeCoin(ctx, in.GetTenantId(), in.GetUserId(), int64(in.GetWalletType()), coin)
			platformBalance, replay = flow.AfterAvailable, true
			return e
		}

		before, e := userAssets.FindOneByTenantIdUserIdWalletTypeCoin(ctx, in.GetTenantId(), in.GetUserId(), int64(in.GetWalletType()), coin)
		if e != nil && e != models.ErrNotFound {
			return e
		}
		if ok, subErr := accounts.SubAvailable(ctx, account.Id, amount, now); subErr != nil {
			return subErr
		} else if !ok {
			return fmt.Errorf("insufficient staking reward platform balance")
		}
		platformBalance = account.AvailableAmount.Sub(amount)
		if before == nil {
			_, e = userAssets.Insert(ctx, &models.TUserAsset{
				TenantId: in.GetTenantId(), UserId: in.GetUserId(), WalletType: int64(in.GetWalletType()), Coin: coin,
				TotalAmount: amount, AvailableAmount: amount, FrozenAmount: decimal.Zero, LockedAmount: decimal.Zero,
				Enabled: 1, Version: 1, Remark: in.GetRemark(), CreateTimes: now, UpdateTimes: now,
			})
		} else {
			_, e = userAssets.AddAvailableAmount(ctx, in.GetTenantId(), in.GetUserId(), int64(in.GetWalletType()), coin, amount, now)
		}
		if e != nil {
			return e
		}
		after, e = userAssets.FindOneByTenantIdUserIdWalletTypeCoin(ctx, in.GetTenantId(), in.GetUserId(), int64(in.GetWalletType()), coin)
		if e != nil {
			return e
		}
		userFlow := helpers.BuildAssetFlowRecord(l.svcCtx, ctx, in.GetTenantId(), in.GetUserId(), int64(in.GetWalletType()), coin, sceneType, bizType, sceneType, in.GetBizId(), bizNo, asset.AssetOpType_ASSET_OP_TYPE_ADD, amount, before, after, in.GetRemark(), now)
		if userFlow == nil {
			return fmt.Errorf("generate asset flow number failed")
		}
		if _, e = assetFlows.Insert(ctx, userFlow); e != nil {
			return e
		}
		if _, e = platformFlows.Insert(ctx, &models.TAssetPlatformFlow{
			TenantId: in.GetTenantId(), PlatformAccountId: account.Id, AccountType: accountType, Coin: coin,
			OpType: 2, Amount: amount, BeforeAvailable: account.AvailableAmount, AfterAvailable: platformBalance,
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
