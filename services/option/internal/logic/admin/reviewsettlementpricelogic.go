package adminlogic

import (
	"context"
	"errors"
	"strings"
	"time"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ReviewSettlementPriceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReviewSettlementPriceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewSettlementPriceLogic {
	return &ReviewSettlementPriceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 确认或拒绝待复核结算价
func (l *ReviewSettlementPriceLogic) ReviewSettlementPrice(in *option.ReviewSettlementPriceReq) (*option.GetSettlementPriceResp, error) {
	operatorID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	reason := strings.TrimSpace(in.Reason)
	if in.TenantId <= 0 || in.SettlementPriceId <= 0 || operatorID <= 0 ||
		len(reason) > 500 || (!in.Approve && reason == "") {
		return &option.GetSettlementPriceResp{
			Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx)),
		}, nil
	}
	now := time.Now().Unix()
	var reviewed *models.TOptionSettlementPrice
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		priceModel := models.NewTOptionSettlementPriceModel(conn, l.svcCtx.Config.CacheRedis)
		item, findErr := priceModel.FindOneForUpdate(ctx, in.SettlementPriceId)
		if findErr != nil {
			return findErr
		}
		if item.TenantId != in.TenantId {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		}
		latest, findErr := priceModel.FindLatestForUpdate(ctx, item.TenantId, item.ContractId)
		if findErr != nil {
			return findErr
		}
		if validationErr := validateSettlementPriceReview(item, latest, operatorID); validationErr != nil {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		}
		if in.Approve {
			if item.SupersedesId > 0 {
				previous, previousErr := priceModel.FindOneForUpdate(ctx, item.SupersedesId)
				if previousErr != nil {
					return previousErr
				}
				if previous.Status != int64(option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_CONFIRMED) {
					return errors.New("settlement price superseded version is not confirmed")
				}
				previous.Status = int64(option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_SUPERSEDED)
				previous.UpdateTimes = now
				if previousErr := priceModel.Update(ctx, previous); previousErr != nil {
					return previousErr
				}
			}
			item.Status = int64(option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_CONFIRMED)
		} else {
			item.Status = int64(option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_REJECTED)
		}
		if reason != "" {
			if item.ChangeReason != "" {
				item.ChangeReason += "; review: " + reason
			} else {
				item.ChangeReason = reason
			}
		}
		item.ConfirmedBy = operatorID
		item.ConfirmedAt = now
		item.UpdateTimes = now
		if updateErr := priceModel.Update(ctx, item); updateErr != nil {
			return updateErr
		}
		reviewed = item
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &option.GetSettlementPriceResp{
		Base: helper.OkResp(), Data: helpers.ToSettlementPriceProto(reviewed),
	}, nil
}

func validateSettlementPriceReview(
	item, latest *models.TOptionSettlementPrice,
	operatorID int64,
) error {
	if item == nil || latest == nil || operatorID <= 0 {
		return errors.New("invalid settlement price review")
	}
	if item.Status != int64(option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_PENDING) {
		return errors.New("settlement price is not pending")
	}
	if latest.Id != item.Id {
		return errors.New("settlement price is not the latest version")
	}
	if item.CreatedBy > 0 && item.CreatedBy == operatorID {
		return errors.New("settlement price requires independent reviewer")
	}
	return nil
}
