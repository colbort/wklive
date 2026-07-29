package adminlogic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/trade"
	"wklive/services/trade/internal/logic/helpers"
	"wklive/services/trade/internal/mapper"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAccountLiquidationDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAccountLiquidationDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountLiquidationDetailLogic {
	return &GetAccountLiquidationDetailLogic{
		ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx),
	}
}

func (l *GetAccountLiquidationDetailLogic) GetAccountLiquidationDetail(
	in *trade.GetAccountLiquidationDetailReq,
) (*trade.GetAccountLiquidationDetailResp, error) {
	tenantID := helpers.AdminTenantID(l.ctx, in.GetTenantId())
	parent, err := l.svcCtx.ContractAccountLiqModel.FindOne(l.ctx, in.GetId())
	if errors.Is(err, models.ErrNotFound) || (err == nil && parent.TenantId != tenantID) {
		return &trade.GetAccountLiquidationDetailResp{
			Base: helper.ErrResp(i18n.BusinessDataNotFound, "account liquidation not found"),
		}, nil
	}
	if err != nil {
		return nil, err
	}
	items, err := l.svcCtx.ContractAccountLiqItemModel.FindByLiquidation(l.ctx, tenantID, parent.Id, false)
	if err != nil {
		return nil, err
	}
	instructions, _, err := l.svcCtx.TradeSettlementInstrModel.FindPage(l.ctx, models.AdminPageFilter{
		TenantId: tenantID, BizType: accountLiquidationBizType, BizId: parent.LiquidationNo,
	}, 0, 100)
	if err != nil {
		return nil, err
	}
	resp := &trade.GetAccountLiquidationDetailResp{
		Base: helper.OkResp(), Data: mapper.AccountLiquidationProto(parent),
	}
	for _, item := range items {
		resp.Items = append(resp.Items, mapper.AccountLiquidationItemProto(item))
		executions, findErr := l.svcCtx.ContractAdlExecutionModel.FindByLiquidation(
			l.ctx, tenantID, -item.Id,
		)
		if findErr != nil {
			return nil, findErr
		}
		for _, execution := range executions {
			if !execution.AssetCredit.IsPositive() {
				continue
			}
			instruction, instructionErr := l.svcCtx.TradeSettlementInstrModel.
				FindOneByTenantIdInstructionNo(l.ctx, tenantID, execution.ExecutionNo)
			if errors.Is(instructionErr, models.ErrNotFound) {
				continue
			}
			if instructionErr != nil {
				return nil, instructionErr
			}
			resp.SettlementInstructions = append(
				resp.SettlementInstructions, mapper.InstructionProto(instruction),
			)
		}
	}
	for _, instruction := range instructions {
		resp.SettlementInstructions = append(resp.SettlementInstructions, mapper.InstructionProto(instruction))
	}
	return resp, nil
}
