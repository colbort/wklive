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

const feeRevenueAccountType = "FEE_REVENUE"

type CreditPlatformRevenueLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreditPlatformRevenueLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreditPlatformRevenueLogic {
	return &CreditPlatformRevenueLogic{ctx: ctx, svcCtx: svcCtx}
}

// CreditPlatformRevenue atomically credits a tenant-owned fee revenue account.
// The platform flow unique key is the business idempotency fence.
func (l *CreditPlatformRevenueLogic) CreditPlatformRevenue(in *asset.CreditPlatformRevenueReq) (*asset.CreditPlatformRevenueResp, error) {
	amount, err := conv.ParseDecimalField(in.GetAmount())
	coin := strings.ToUpper(strings.TrimSpace(in.GetCoin()))
	bizNo := strings.TrimSpace(in.GetBizNo())
	bizType := helpers.AssetBizType(in.GetBizType())
	sceneType := helpers.AssetSceneType(in.GetSceneType())
	if err != nil || !amount.IsPositive() || in.GetTenantId() <= 0 || coin == "" || bizNo == "" || bizType == "" || sceneType == "" {
		return nil, fmt.Errorf("invalid platform revenue request")
	}

	var accountID int64
	var balance decimal.Decimal
	replay := false
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		accounts := models.NewTAssetPlatformAccountModel(conn, l.svcCtx.Config.CacheRedis)
		flows := models.NewTAssetPlatformFlowModel(conn, l.svcCtx.Config.CacheRedis)

		account, findErr := accounts.FindOneForUpdate(ctx, in.GetTenantId(), feeRevenueAccountType, coin)
		if findErr != nil {
			return fmt.Errorf("fee revenue platform account is not configured: %w", findErr)
		}
		accountID, balance = account.Id, account.AvailableAmount

		existing, findErr := flows.FindOneByTenantIdPlatformAccountIdSceneTypeBizNo(ctx, in.GetTenantId(), account.Id, sceneType, bizNo)
		if findErr == nil {
			if !platformRevenueFlowMatches(existing, amount, bizType, sceneType, in.GetBizId()) {
				return fmt.Errorf("platform revenue idempotency parameters changed")
			}
			replay = true
			return nil
		}
		if findErr != models.ErrNotFound {
			return findErr
		}

		now := utils.NowMillis()
		if err = accounts.AddAvailable(ctx, account.Id, amount, now); err != nil {
			return err
		}
		balance = account.AvailableAmount.Add(amount)
		_, err = flows.Insert(ctx, &models.TAssetPlatformFlow{
			TenantId:          in.GetTenantId(),
			PlatformAccountId: account.Id,
			AccountType:       account.AccountType,
			Coin:              coin,
			OpType:            1,
			Amount:            amount,
			BeforeAvailable:   account.AvailableAmount,
			AfterAvailable:    balance,
			BizType:           bizType,
			SceneType:         sceneType,
			BizId:             in.GetBizId(),
			BizNo:             bizNo,
			Remark:            in.GetRemark(),
			CreateTimes:       now,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return &asset.CreditPlatformRevenueResp{
		Base:                   helper.OkResp(),
		PlatformAccountId:      accountID,
		PlatformAccountBalance: balance.String(),
		IdempotentReplay:       replay,
	}, nil
}

func platformRevenueFlowMatches(flow *models.TAssetPlatformFlow, amount decimal.Decimal, bizType, sceneType string, bizID int64) bool {
	return flow != nil &&
		flow.AccountType == feeRevenueAccountType &&
		flow.OpType == 1 &&
		flow.Amount.Equal(amount) &&
		flow.BizType == bizType &&
		flow.SceneType == sceneType &&
		flow.BizId == bizID
}
