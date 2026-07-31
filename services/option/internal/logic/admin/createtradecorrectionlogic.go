package adminlogic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"wklive/common/generate"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/option"
	applogic "wklive/services/option/internal/logic/app"
	"wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CreateTradeCorrectionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTradeCorrectionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTradeCorrectionLogic {
	return &CreateTradeCorrectionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建异常成交现金更正草案，同时暂停合约并撤销活动订单
func (l *CreateTradeCorrectionLogic) CreateTradeCorrection(in *option.CreateTradeCorrectionReq) (*option.GetTradeCorrectionResp, error) {
	operatorID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	reason, evidenceRef := strings.TrimSpace(in.Reason), strings.TrimSpace(in.EvidenceRef)
	if in.TenantId <= 0 || in.TradeId <= 0 || operatorID <= 0 ||
		in.Action != option.TradeCorrectionAction_TRADE_CORRECTION_ACTION_CASH_ADJUSTMENT ||
		reason == "" || evidenceRef == "" || len(reason) > 500 || len(evidenceRef) > 500 {
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
	trade, err := helpers.FindTradeByNoOrID(l.ctx, l.svcCtx, in.TenantId, in.TradeId, "")
	if err != nil {
		return nil, err
	}
	contract, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, trade.ContractId)
	if err != nil {
		return nil, err
	}
	legs, err := validateTradeCorrectionLegs(in.Legs, trade, contract)
	if err != nil {
		return &option.GetTradeCorrectionResp{
			Base: helper.ErrResp(i18n.ParamError, err.Error()),
		}, nil
	}
	caseNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "order_id", "OTC", "")
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	var created *models.TOptionTradeCorrection
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		contractModel := models.NewTOptionContractModel(conn, l.svcCtx.Config.CacheRedis)
		correctionModel := models.NewTOptionTradeCorrectionModel(conn, l.svcCtx.Config.CacheRedis)
		legModel := models.NewTOptionTradeCorrectionLegModel(conn, l.svcCtx.Config.CacheRedis)
		eventModel := models.NewTOptionTradingControlEventModel(conn, l.svcCtx.Config.CacheRedis)

		lockedContract, findErr := contractModel.FindOneForUpdate(ctx, trade.ContractId)
		if findErr != nil {
			return findErr
		}
		if lockedContract.TenantId != in.TenantId {
			return errInvalidTradeCorrection
		}
		if _, findErr = correctionModel.FindActiveByTradeForUpdate(
			ctx, in.TenantId, trade.Id,
		); findErr == nil {
			return errActiveTradeCorrection
		} else if !errors.Is(findErr, models.ErrNotFound) {
			return findErr
		}
		if lockedContract.Status != int64(option.ContractStatus_CONTRACT_STATUS_PAUSED) {
			previousStatus := lockedContract.Status
			lockedContract.Status = int64(option.ContractStatus_CONTRACT_STATUS_PAUSED)
			lockedContract.UpdateTimes = now
			if updateErr := contractModel.Update(ctx, lockedContract); updateErr != nil {
				return updateErr
			}
			if _, insertErr := eventModel.Insert(ctx, &models.TOptionTradingControlEvent{
				TenantId: in.TenantId, ContractId: trade.ContractId,
				EventType: "CONTRACT_STATUS_UPDATED", Reason: "ERRONEOUS_TRADE_REPORTED",
				Detail: fmt.Sprintf(
					"status:%d->%d trade_id=%d case_no=%s",
					previousStatus, lockedContract.Status, trade.Id, caseNo,
				),
				OperatorId: operatorID, CreateTimes: now,
			}); insertErr != nil {
				return insertErr
			}
		}
		created = &models.TOptionTradeCorrection{
			TenantId: in.TenantId, CaseNo: caseNo, TradeId: trade.Id,
			ContractId: trade.ContractId, Action: int64(in.Action),
			Status: int64(option.TradeCorrectionStatus_TRADE_CORRECTION_STATUS_PENDING_REVIEW),
			Reason: reason, EvidenceRef: evidenceRef, RequestedBy: operatorID,
			CreateTimes: now, UpdateTimes: now,
		}
		result, insertErr := correctionModel.Insert(ctx, created)
		if insertErr != nil {
			return insertErr
		}
		created.Id, insertErr = result.LastInsertId()
		if insertErr != nil {
			return insertErr
		}
		for index, leg := range legs {
			legNo := int64(index + 1)
			if _, insertErr = legModel.Insert(ctx, &models.TOptionTradeCorrectionLeg{
				TenantId: in.TenantId, CorrectionId: created.Id, LegNo: legNo,
				UserId: leg.UserID, AccountId: leg.AccountID, Coin: leg.Coin,
				Direction: int64(leg.Direction), Amount: leg.Amount,
				InstructionNo: fmt.Sprintf("%s-LEG-%03d", caseNo, legNo),
				CreateTimes:   now,
			}); insertErr != nil {
				return insertErr
			}
		}
		_, insertErr = eventModel.Insert(ctx, &models.TOptionTradingControlEvent{
			TenantId: in.TenantId, ContractId: trade.ContractId,
			EventType: "TRADE_CORRECTION_CREATED", Reason: "ERRONEOUS_TRADE",
			Detail: fmt.Sprintf(
				"case_no=%s trade_id=%d action=%d evidence=%s",
				caseNo, trade.Id, in.Action, evidenceRef,
			),
			OperatorId: operatorID, CreateTimes: now,
		})
		return insertErr
	})
	if errors.Is(err, errActiveTradeCorrection) || errors.Is(err, errInvalidTradeCorrection) {
		return &option.GetTradeCorrectionResp{
			Base: helper.ErrResp(i18n.OperationNotAllowed, err.Error()),
		}, nil
	}
	if err != nil {
		return nil, err
	}
	if cancelErr := applogic.CancelContractOrdersByControl(
		l.ctx, l.svcCtx, in.TenantId, trade.ContractId, "ERRONEOUS_TRADE_REPORTED",
	); cancelErr != nil {
		l.Errorf(
			"cancel contract orders after trade correction failed, caseNo=%s err=%v",
			caseNo, cancelErr,
		)
		created.LastErrorMsg = cancelErr.Error()
		created.UpdateTimes = time.Now().Unix()
		if len(created.LastErrorMsg) > 500 {
			created.LastErrorMsg = created.LastErrorMsg[:500]
		}
		if updateErr := l.svcCtx.OptionTradeCorrectionModel.Update(l.ctx, created); updateErr != nil {
			return nil, errors.Join(cancelErr, updateErr)
		}
	}
	storedLegs, err := l.svcCtx.OptionTradeCorrectionLegModel.FindByCorrection(
		l.ctx, created.TenantId, created.Id,
	)
	if err != nil {
		return nil, err
	}
	return &option.GetTradeCorrectionResp{
		Base: helper.OkResp(), Data: helpers.ToTradeCorrectionProto(created, storedLegs),
	}, nil
}
