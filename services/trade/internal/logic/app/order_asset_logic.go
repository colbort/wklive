package applogic

import (
	"context"
	"errors"
	"fmt"
	helpers "wklive/services/trade/internal/logic/helpers"

	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// assetFreezeError distinguishes an explicit Asset rejection from an RPC
// outcome that is unknown to Trade. Only an explicit response is safe to map
// to ORDER_STATUS_REJECTED.
type assetFreezeError struct {
	err        error
	definitive bool
}

func (e *assetFreezeError) Error() string { return e.err.Error() }
func (e *assetFreezeError) Unwrap() error { return e.err }

func isDefinitiveAssetFreezeError(err error) bool {
	var target *assetFreezeError
	return errors.As(err, &target) && target.definitive
}

func freezeOrderAsset(
	svcCtx *svc.ServiceContext,
	ctx context.Context,
	order *models.TTradeOrder,
	symbol *models.TTradeSymbol,
	frozenAsset string,
	frozenAmount decimal.Decimal,
) (string, error) {
	if order == nil || symbol == nil || frozenAsset == "" || !frozenAmount.IsPositive() {
		return "", nil
	}

	resp, err := svcCtx.AssetClient.FreezeAsset(ctx, &asset.FreezeAssetReq{
		TenantId:   order.TenantId,
		UserId:     order.UserId,
		WalletType: helpers.WalletTypeForProduct(common.ProductType(order.ProductType)),
		Coin:       frozenAsset,
		Amount:     frozenAmount.String(),
		BizType:    asset.BizType_BIZ_TYPE_TRADE,
		SceneType:  asset.SceneType_SCENE_TYPE_PLACE_ORDER,
		BizId:      order.Id,
		BizNo:      order.OrderNo,
		Remark:     "trade place order freeze",
	})
	if err != nil {
		return "", &assetFreezeError{err: freezeAssetContextError(order.UserId, helpers.WalletTypeForProduct(common.ProductType(order.ProductType)), frozenAsset, err)}
	}
	if resp == nil || resp.Base == nil {
		return "", &assetFreezeError{err: freezeAssetContextError(order.UserId, helpers.WalletTypeForProduct(common.ProductType(order.ProductType)), frozenAsset, i18n.StatusError(ctx, i18n.InternalServerError))}
	}
	if resp.Base.Code != 200 {
		return "", &assetFreezeError{err: freezeAssetContextError(order.UserId, helpers.WalletTypeForProduct(common.ProductType(order.ProductType)), frozenAsset, i18n.StatusError(ctx, resp.Base.Code)), definitive: true}
	}

	return resp.GetData().GetFreezeNo(), nil
}

func freezeAssetContextError(userID int64, walletType common.WalletType, coin string, err error) error {
	return fmt.Errorf("userId=%d walletType=%s(%d) coin=%s: %w", userID, walletType.String(), walletType, coin, err)
}

func unfreezeRemainingOrderAsset(svcCtx *svc.ServiceContext, ctx context.Context, order *models.TTradeOrder, reason string) error {
	if order == nil || order.OrderNo == "" {
		return nil
	}
	var instructionID int64
	now := utils.NowMillis()
	err := svcCtx.DB.TransactCtx(ctx, func(txCtx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		reservationModel := models.NewTTradeAssetReservationModel(conn, svcCtx.Config.CacheRedis)
		instructionModel := models.NewTTradeSettlementInstructionModel(conn, svcCtx.Config.CacheRedis)
		reservation, err := reservationModel.FindOneByReservationNoForUpdate(txCtx, order.TenantId, order.OrderNo)
		if errors.Is(err, models.ErrNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		unfinished, err := instructionModel.CountUnfinishedByOrder(txCtx, order.TenantId, order.Id)
		if err != nil {
			return err
		}
		if unfinished > 0 {
			return nil
		}
		remaining := reservation.ReservedAmount.Sub(reservation.ConsumedAmount).Sub(reservation.ReleasedAmount)
		if !remaining.IsPositive() {
			return nil
		}
		instructionNo := derivedTradeBizNo(order.OrderNo, "RELEASE")
		if err := insertSettlementInstructionIdempotent(txCtx, instructionModel, &models.TTradeSettlementInstruction{TenantId: order.TenantId, InstructionNo: instructionNo, BizType: "order", BizId: order.OrderNo, OrderId: order.Id, ReservationNo: order.OrderNo, UserId: order.UserId, Action: int64(trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_RELEASE_FROZEN), Asset: reservation.Asset, Amount: remaining, StepNo: 1, Status: int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_PENDING), NextRetryAt: now, LastErrorMsg: reason, CreateTimes: now, UpdateTimes: now}); err != nil {
			return err
		}
		instruction, err := instructionModel.FindOneByTenantIdInstructionNo(txCtx, order.TenantId, instructionNo)
		if err != nil {
			return err
		}
		instructionID = instruction.Id
		_, err = reservationModel.BeginRelease(txCtx, reservation.Id, now)
		return err
	})
	if err != nil {
		return err
	}
	if instructionID == 0 {
		return svcCtx.DB.TransactCtx(ctx, func(txCtx context.Context, session sqlx.Session) error {
			_, err := finalizeOrderTermination(txCtx, sqlx.NewSqlConnFromSession(session), svcCtx, order.Id, utils.NowMillis())
			return err
		})
	}
	return NewProcessReservationReleasesLogic(ctx, svcCtx).ProcessInstruction(instructionID)
}
