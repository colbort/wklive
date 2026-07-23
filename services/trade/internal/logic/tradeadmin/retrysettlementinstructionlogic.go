package tradeadminlogic

import (
	"context"
	"errors"

	"wklive/common/generate"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type RetrySettlementInstructionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRetrySettlementInstructionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetrySettlementInstructionLogic {
	return &RetrySettlementInstructionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 仅重置失败/人工处理的结算指令；不得修改金额
func (l *RetrySettlementInstructionLogic) RetrySettlementInstruction(in *trade.RetrySettlementInstructionReq) (*trade.AdminCommonResp, error) {
	tenantID := adminTenantID(l.ctx, in.TenantId)
	operatorID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil || operatorID <= 0 {
		return &trade.AdminCommonResp{Base: helper.ErrResp(i18n.OperationNotAllowed, "missing admin operator identity")}, nil
	}
	eventNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "order_id", "TRE", "")
	if err != nil {
		return nil, err
	}
	notFound, invalidStatus := false, false
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		notFound, invalidStatus, err = retrySettlementInstructionTx(ctx, sqlx.NewSqlConnFromSession(session), l.svcCtx, in, tenantID, operatorID, eventNo, utils.NowMillis())
		return err
	})
	if err != nil {
		return nil, err
	}
	if notFound {
		return &trade.AdminCommonResp{Base: helper.ErrResp(i18n.BusinessDataNotFound, "settlement instruction not found")}, nil
	}
	if invalidStatus {
		return &trade.AdminCommonResp{Base: helper.ErrResp(i18n.OperationNotAllowed, "only failed or manual-review instructions can be retried")}, nil
	}
	return &trade.AdminCommonResp{Base: helper.OkResp()}, nil
}

func retrySettlementInstructionTx(ctx context.Context, conn sqlx.SqlConn, svcCtx *svc.ServiceContext, in *trade.RetrySettlementInstructionReq, tenantID, operatorID int64, eventNo string, now int64) (bool, bool, error) {
	instructionModel := models.NewTTradeSettlementInstructionModel(conn, svcCtx.Config.CacheRedis)
	item, err := instructionModel.FindOneForUpdate(ctx, in.Id)
	if errors.Is(err, models.ErrNotFound) || (err == nil && item.TenantId != tenantID) {
		return true, false, nil
	}
	if err != nil {
		return false, false, err
	}
	if item.Status != int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_FAILED) && item.Status != int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_MANUAL_REVIEW) {
		return false, true, nil
	}
	item.Status, item.NextRetryAt, item.LastErrorMsg, item.UpdateTimes = int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_PENDING), now, "", now
	if err = instructionModel.Update(ctx, item); err != nil {
		return false, false, err
	}
	_, err = models.NewTBizTradeEventModel(conn, svcCtx.Config.CacheRedis).Insert(ctx, &models.TBizTradeEvent{TenantId: tenantID, EventNo: eventNo, EventType: "SETTLEMENT_INSTRUCTION_RETRY_REQUESTED", BizId: item.InstructionNo, BizType: "settlement_instruction", UserId: item.UserId, OperatorId: operatorID, Source: int64(trade.SourceType_SOURCE_TYPE_ADMIN), EventStatus: int64(trade.EventStatus_EVENT_STATUS_PENDING), MaxRetryCount: 20, NextRetryAt: now, Payload: normalizeTradeEventJSON(in.Reason), CreateTimes: now, UpdateTimes: now})
	return false, false, err
}
