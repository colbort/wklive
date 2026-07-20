package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
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
	item, err := l.svcCtx.TradeSettlementInstrModel.FindOne(l.ctx, in.Id)
	if errors.Is(err, models.ErrNotFound) || (err == nil && item.TenantId != tenantID) {
		return &trade.AdminCommonResp{Base: helper.ErrResp(i18n.BusinessDataNotFound, "settlement instruction not found")}, nil
	}
	if err != nil {
		return nil, err
	}
	if item.Status != 4 && item.Status != 5 {
		return &trade.AdminCommonResp{Base: helper.ErrResp(i18n.OperationNotAllowed, "only failed or manual-review instructions can be retried")}, nil
	}
	item.Status = 1
	item.NextRetryAt = utils.NowMillis()
	item.LastErrorMsg = ""
	item.UpdateTimes = utils.NowMillis()
	if err = l.svcCtx.TradeSettlementInstrModel.Update(l.ctx, item); err != nil {
		return nil, err
	}
	eventNo, err := l.svcCtx.GenerateBizNo(l.ctx, "TRE")
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.BizTradeEventModel.Insert(l.ctx, &models.TBizTradeEvent{TenantId: tenantID, EventNo: eventNo, EventType: "SETTLEMENT_INSTRUCTION_RETRY_REQUESTED", BizId: item.InstructionNo, BizType: "settlement_instruction", UserId: item.UserId, OperatorId: operatorID, Source: int64(trade.SourceType_SOURCE_TYPE_ADMIN), EventStatus: int64(trade.EventStatus_EVENT_STATUS_PENDING), MaxRetryCount: 20, NextRetryAt: utils.NowMillis(), Payload: normalizeTradeEventJSON(in.Reason), CreateTimes: utils.NowMillis(), UpdateTimes: utils.NowMillis()})
	if err != nil {
		return nil, err
	}
	return &trade.AdminCommonResp{Base: helper.OkResp()}, nil
}
