package assetinternallogic

import (
	"context"
	"fmt"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ReverseInsuranceCoverLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReverseInsuranceCoverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReverseInsuranceCoverLogic {
	return &ReverseInsuranceCoverLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *ReverseInsuranceCoverLogic) ReverseInsuranceCover(in *asset.ReverseInsuranceCoverReq) (*asset.ChangeAssetResp, error) {
	if in.GetTenantId() <= 0 || in.GetLiquidationNo() == "" || in.GetReversalNo() == "" {
		return nil, fmt.Errorf("invalid insurance reversal request")
	}
	now := utils.NowMillis()
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		covers := models.NewTAssetInsuranceCoverModel(conn, l.svcCtx.Config.CacheRedis)
		accounts := models.NewTAssetPlatformAccountModel(conn, l.svcCtx.Config.CacheRedis)
		flows := models.NewTAssetPlatformFlowModel(conn, l.svcCtx.Config.CacheRedis)
		idempotent := models.NewTAssetIdempotentModel(conn, l.svcCtx.Config.CacheRedis)
		cover, err := covers.FindOneForUpdate(ctx, in.GetTenantId(), in.GetLiquidationNo())
		if err != nil {
			return err
		}
		done, err := prepareAssetIdempotent(ctx, idempotent, in.GetTenantId(), assetBizType(asset.BizType_BIZ_TYPE_INSURANCE_FUND), assetSceneType(asset.SceneType_SCENE_TYPE_INSURANCE_FUND_REVERSAL), in.GetReversalNo(), in.GetRemark(), now)
		if err != nil {
			return err
		}
		if cover.Status == 2 {
			if done {
				return nil
			}
			return fmt.Errorf("insurance cover already reversed by a different reversal number")
		}
		if done {
			return fmt.Errorf("insurance reversal idempotency completed before cover was marked reversed")
		}
		account, err := accounts.FindOneForUpdate(ctx, cover.TenantId, insuranceFundAccountType, cover.Coin)
		if err != nil || account.Id != cover.PlatformAccountId {
			return fmt.Errorf("insurance platform account changed")
		}
		if cover.CoveredAmount.IsPositive() {
			if err = accounts.AddAvailable(ctx, account.Id, cover.CoveredAmount, now); err != nil {
				return err
			}
			_, err = flows.Insert(ctx, &models.TAssetPlatformFlow{TenantId: cover.TenantId, PlatformAccountId: account.Id, AccountType: account.AccountType, Coin: cover.Coin, OpType: 1, Amount: cover.CoveredAmount, BeforeAvailable: account.AvailableAmount, AfterAvailable: account.AvailableAmount.Add(cover.CoveredAmount), BizType: assetBizType(asset.BizType_BIZ_TYPE_INSURANCE_FUND), SceneType: assetSceneType(asset.SceneType_SCENE_TYPE_INSURANCE_FUND_REVERSAL), BizId: cover.LiquidationId, BizNo: in.GetReversalNo(), Remark: in.GetRemark(), CreateTimes: now})
			if err != nil {
				return err
			}
		}
		if err = covers.MarkReversed(ctx, cover.Id, now); err != nil {
			return err
		}
		return completeAssetIdempotent(ctx, idempotent, in.GetTenantId(), assetBizType(asset.BizType_BIZ_TYPE_INSURANCE_FUND), assetSceneType(asset.SceneType_SCENE_TYPE_INSURANCE_FUND_REVERSAL), in.GetReversalNo(), now)
	})
	if err != nil {
		return nil, err
	}
	return &asset.ChangeAssetResp{Base: helper.OkResp(), Data: &asset.ChangeAssetData{BizNo: in.GetReversalNo()}}, nil
}
