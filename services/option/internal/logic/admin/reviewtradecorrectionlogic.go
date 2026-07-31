package adminlogic

import (
	"context"
	"errors"
	"fmt"
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

type ReviewTradeCorrectionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReviewTradeCorrectionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewTradeCorrectionLogic {
	return &ReviewTradeCorrectionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 由独立管理员批准或拒绝异常成交更正
func (l *ReviewTradeCorrectionLogic) ReviewTradeCorrection(in *option.ReviewTradeCorrectionReq) (*option.GetTradeCorrectionResp, error) {
	operatorID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	reason := strings.TrimSpace(in.Reason)
	if in.TenantId <= 0 || in.CorrectionId <= 0 || operatorID <= 0 ||
		reason == "" || len(reason) > 500 {
		return &option.GetTradeCorrectionResp{
			Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx)),
		}, nil
	}
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &option.GetTradeCorrectionResp{
			Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx)),
		}, nil
	}
	now := time.Now().Unix()
	var reviewed *models.TOptionTradeCorrection
	var reviewedLegs []*models.TOptionTradeCorrectionLeg
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		correctionModel := models.NewTOptionTradeCorrectionModel(conn, l.svcCtx.Config.CacheRedis)
		legModel := models.NewTOptionTradeCorrectionLegModel(conn, l.svcCtx.Config.CacheRedis)
		instructionModel := models.NewTOptionAssetInstructionModel(conn, l.svcCtx.Config.CacheRedis)
		eventModel := models.NewTOptionTradingControlEventModel(conn, l.svcCtx.Config.CacheRedis)

		item, findErr := correctionModel.FindOneForUpdate(ctx, in.CorrectionId)
		if findErr != nil {
			return findErr
		}
		if item.TenantId != in.TenantId ||
			item.Status != int64(option.TradeCorrectionStatus_TRADE_CORRECTION_STATUS_PENDING_REVIEW) ||
			item.RequestedBy == operatorID {
			return errInvalidTradeCorrection
		}
		legs, findErr := legModel.FindByCorrection(ctx, item.TenantId, item.Id)
		if findErr != nil {
			return findErr
		}
		if len(legs) < 2 {
			return errInvalidTradeCorrection
		}
		item.ReviewedBy = operatorID
		item.ReviewReason = reason
		item.ReviewedAt = now
		item.UpdateTimes = now
		eventType := "TRADE_CORRECTION_REJECTED"
		if !in.Approve {
			item.Status = int64(option.TradeCorrectionStatus_TRADE_CORRECTION_STATUS_REJECTED)
		} else {
			eventType = "TRADE_CORRECTION_APPROVED"
			item.Status = int64(option.TradeCorrectionStatus_TRADE_CORRECTION_STATUS_EXECUTING)
			for _, leg := range legs {
				action := option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEBIT_AVAILABLE
				stepNo := int64(1)
				if leg.Direction == int64(option.TradeCorrectionLegDirection_TRADE_CORRECTION_LEG_DIRECTION_CREDIT) {
					action = option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_CREDIT_AVAILABLE
					stepNo = 2
				} else if leg.Direction != int64(option.TradeCorrectionLegDirection_TRADE_CORRECTION_LEG_DIRECTION_DEBIT) {
					return errInvalidTradeCorrection
				}
				if _, insertErr := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
					TenantId: item.TenantId, InstructionNo: leg.InstructionNo,
					BizNo: item.CaseNo, TradeId: item.TradeId,
					UserId: leg.UserId, AccountId: leg.AccountId,
					Action: int64(action), Coin: leg.Coin, Amount: leg.Amount,
					StepNo:               stepNo,
					Status:               int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
					ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
					CreateTimes:          now, UpdateTimes: now,
				}); insertErr != nil {
					return insertErr
				}
			}
		}
		if updateErr := correctionModel.Update(ctx, item); updateErr != nil {
			return updateErr
		}
		if _, insertErr := eventModel.Insert(ctx, &models.TOptionTradingControlEvent{
			TenantId: item.TenantId, ContractId: item.ContractId,
			EventType: eventType, Reason: "INDEPENDENT_REVIEW",
			Detail: fmt.Sprintf(
				"case_no=%s trade_id=%d approve=%t reason=%s",
				item.CaseNo, item.TradeId, in.Approve, reason,
			),
			OperatorId: operatorID, CreateTimes: now,
		}); insertErr != nil {
			return insertErr
		}
		reviewed, reviewedLegs = item, legs
		return nil
	})
	if errors.Is(err, errInvalidTradeCorrection) {
		return &option.GetTradeCorrectionResp{
			Base: helper.ErrResp(i18n.OperationNotAllowed, err.Error()),
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &option.GetTradeCorrectionResp{
		Base: helper.OkResp(), Data: helpers.ToTradeCorrectionProto(reviewed, reviewedLegs),
	}, nil
}
