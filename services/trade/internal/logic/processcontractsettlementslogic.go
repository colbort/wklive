package logic

import (
	"context"

	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProcessContractSettlementsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProcessContractSettlementsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessContractSettlementsLogic {
	return &ProcessContractSettlementsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 合约结算（资金费率/交割合约/秒合约）
func (l *ProcessContractSettlementsLogic) ProcessContractSettlements(in *trade.TradeTaskReq) (*trade.TradeTaskResp, error) {
	return runTradeTaskWithLock(l.ctx, l.svcCtx, "process_contract_settlements", func() (*trade.TradeTaskResp, error) {
		if err := l.settleFundingFees(in); err != nil {
			return nil, err
		}
		if err := l.disableExpiredSymbols(in.GetTenantId(), trade.ContractType_CONTRACT_TYPE_DELIVERY); err != nil {
			return nil, err
		}
		if err := l.disableExpiredSymbols(in.GetTenantId(), trade.ContractType_CONTRACT_TYPE_SECONDS); err != nil {
			return nil, err
		}
		return okTradeTaskResp(), nil
	})
}

func (l *ProcessContractSettlementsLogic) settleFundingFees(in *trade.TradeTaskReq) error {
	now := utils.NowMillis()
	cursor := int64(0)
	for {
		contracts, _, err := l.svcCtx.TradeSymbolContractModel.FindPage(l.ctx, cursor, 100)
		if err != nil {
			return err
		}
		if len(contracts) == 0 {
			return nil
		}
		for _, contract := range contracts {
			cursor = contract.Id
			if in.GetTenantId() > 0 && contract.TenantId != in.GetTenantId() {
				continue
			}
			if contract.FundingIntervalMinutes <= 0 {
				continue
			}
			symbol, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, contract.SymbolId)
			if err != nil {
				return err
			}
			if symbol.ContractType != int64(trade.ContractType_CONTRACT_TYPE_PERPETUAL) {
				continue
			}
			intervalMillis := contract.FundingIntervalMinutes * 60 * 1000
			if intervalMillis <= 0 || now%intervalMillis > 60*1000 {
				continue
			}
			if err := createTradeTaskEvent(l.ctx, l.svcCtx, contract.TenantId, "FUNDING_FEE_SETTLEMENT_REQUIRED", "symbol", symbol.Id, 0, symbol.Id, symbol.MarketType, "funding fee task"); err != nil {
				return err
			}
		}
		if len(contracts) < 100 {
			return nil
		}
	}
}

func (l *ProcessContractSettlementsLogic) disableExpiredSymbols(tenantID int64, contractType trade.ContractType) error {
	now := utils.NowMillis()
	cursor := int64(0)
	for {
		contracts, _, err := l.svcCtx.TradeSymbolContractModel.FindPage(l.ctx, cursor, 100)
		if err != nil {
			return err
		}
		if len(contracts) == 0 {
			return nil
		}
		for _, contract := range contracts {
			cursor = contract.Id
			if tenantID > 0 && contract.TenantId != tenantID {
				continue
			}
			if contract.DeliveryTime == 0 || contract.DeliveryTime > now {
				continue
			}
			symbol, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, contract.SymbolId)
			if err != nil {
				return err
			}
			if symbol.ContractType != int64(contractType) || symbol.Status == int64(trade.SymbolStatus_SYMBOL_STATUS_DISABLED) {
				continue
			}
			symbol.Status = int64(trade.SymbolStatus_SYMBOL_STATUS_DISABLED)
			symbol.UpdateTimes = now
			if err := l.svcCtx.TradeSymbolModel.Update(l.ctx, symbol); err != nil {
				return err
			}
			eventNo, err := l.svcCtx.GenerateBizNo(l.ctx, "TRE")
			if err != nil {
				return err
			}
			if _, err := l.svcCtx.BizTradeEventModel.Insert(l.ctx, &models.TBizTradeEvent{
				TenantId:      symbol.TenantId,
				EventNo:       eventNo,
				EventType:     "CONTRACT_SETTLED",
				BizId:         symbol.Symbol,
				BizType:       "symbol",
				SymbolId:      symbol.Id,
				MarketType:    symbol.MarketType,
				Source:        int64(trade.SourceType_SOURCE_TYPE_TASK),
				EventStatus:   int64(trade.EventStatus_EVENT_STATUS_PENDING),
				MaxRetryCount: 3,
				NextRetryAt:   now,
				Payload:       symbol.Symbol,
				CreateTimes:   now,
				UpdateTimes:   now,
			}); err != nil {
				return err
			}
		}
		if len(contracts) < 100 {
			return nil
		}
	}
}
