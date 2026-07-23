package adminlogic

import (
	"context"
	"fmt"
	"strings"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AdjustPlatformAccountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdjustPlatformAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdjustPlatformAccountLogic {
	return &AdjustPlatformAccountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 幂等调整平台账户余额
func (l *AdjustPlatformAccountLogic) AdjustPlatformAccount(in *asset.AdjustPlatformAccountReq) (*asset.PlatformAccountResp, error) {
	typeName, coin := normalizePlatformAccount(in.GetAccountType(), in.GetCoin())
	requestNo := strings.TrimSpace(in.GetRequestNo())
	amount, err := decimal.NewFromString(in.GetAmount())
	if err != nil || !amount.IsPositive() || in.GetTenantId() <= 0 || typeName != insuranceFundAccountType || coin == "" || requestNo == "" || (in.GetDirection() != 1 && in.GetDirection() != 2) {
		return nil, fmt.Errorf("invalid platform account adjustment")
	}
	var result *models.TAssetPlatformAccount
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		accounts := models.NewTAssetPlatformAccountModel(conn, l.svcCtx.Config.CacheRedis)
		flows := models.NewTAssetPlatformFlowModel(conn, l.svcCtx.Config.CacheRedis)
		account, e := accounts.FindOneForUpdate(ctx, in.GetTenantId(), typeName, coin)
		if e != nil {
			return e
		}
		flow, e := flows.FindOneByTenantIdPlatformAccountIdSceneTypeBizNo(ctx, in.GetTenantId(), account.Id, "platform_manual_adjust", requestNo)
		if e == nil {
			if flow.OpType != int64(in.GetDirection()) || !flow.Amount.Equal(amount) {
				return fmt.Errorf("request_no reused with different adjustment")
			}
			account.AvailableAmount = flow.AfterAvailable
			result = account
			return nil
		}
		if e != models.ErrNotFound {
			return e
		}
		before := account.AvailableAmount
		now := utils.NowMillis()
		if in.GetDirection() == 1 {
			e = accounts.AddAvailable(ctx, account.Id, amount, now)
			account.AvailableAmount = before.Add(amount)
		} else {
			var ok bool
			ok, e = accounts.SubAvailable(ctx, account.Id, amount, now)
			if e == nil && !ok {
				return fmt.Errorf("insufficient platform account balance")
			}
			account.AvailableAmount = before.Sub(amount)
		}
		if e != nil {
			return e
		}
		_, e = flows.Insert(ctx, &models.TAssetPlatformFlow{TenantId: in.GetTenantId(), PlatformAccountId: account.Id, AccountType: typeName, Coin: coin, OpType: int64(in.GetDirection()), Amount: amount, BeforeAvailable: before, AfterAvailable: account.AvailableAmount, BizType: "admin", SceneType: "platform_manual_adjust", BizNo: requestNo, Remark: in.GetRemark(), CreateTimes: now})
		if e != nil {
			return e
		}
		account.Version++
		account.UpdateTimes = now
		result = account
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &asset.PlatformAccountResp{Base: helper.OkResp(), Data: platformAccountProto(result)}, nil
}
