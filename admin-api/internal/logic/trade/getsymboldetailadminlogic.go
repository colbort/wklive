// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"context"
	"errors"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/common/utils"
	"wklive/proto/trade"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSymbolDetailAdminLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSymbolDetailAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSymbolDetailAdminLogic {
	return &GetSymbolDetailAdminLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSymbolDetailAdminLogic) GetSymbolDetailAdmin(req *types.GetSymbolDetailAdminReq) (resp *types.GetSymbolDetailAdminResp, err error) {
	resp, err = logicutil.Proxy[types.GetSymbolDetailAdminResp](l.ctx, req, l.svcCtx.TradeCli.GetSymbolDetailAdmin)
	if err != nil || resp == nil || resp.Code != 200 {
		if err != nil {
			return logicutil.SystemErrorResp[types.GetSymbolDetailAdminResp](l.ctx, err)
		}
		if resp == nil {
			return logicutil.SystemErrorResp[types.GetSymbolDetailAdminResp](l.ctx, errors.New("empty symbol detail response"))
		}
		return resp, nil
	}

	tenantId := req.TenantId
	if tenantId == 0 {
		tenantId = resp.Data.Symbol.TenantId
	}
	ctx := context.WithValue(l.ctx, utils.CtxKeyTenantId, tenantId)
	detail, err := logicutil.Proxy[trade.GetSymbolDetailResp](ctx, &types.GetSymbolDetailAdminReq{Id: req.Id}, l.svcCtx.TradeAppCli.GetSymbolDetail)
	if err != nil || detail == nil || detail.Base.Code != 200 {
		if err != nil {
			return logicutil.SystemErrorResp[types.GetSymbolDetailAdminResp](l.ctx, err)
		}
		if detail == nil {
			return logicutil.SystemErrorResp[types.GetSymbolDetailAdminResp](l.ctx, errors.New("empty app symbol detail response"))
		}
		resp.Code = detail.GetBase().GetCode()
		resp.Msg = detail.GetBase().GetMsg()
		return resp, nil
	}
	if detail.GetData().GetSpot() != nil {
		resp.Data.Spot = tradeSymbolSpotToTypes(detail.GetData().GetSpot())
	}
	if detail.GetData().GetContract() != nil {
		resp.Data.Contract = tradeSymbolContractToTypes(detail.GetData().GetContract())
	}
	return resp, nil
}

func tradeSymbolSpotToTypes(item *trade.TradeSymbolSpot) types.TradeSymbolSpot {
	return types.TradeSymbolSpot{
		Id:           item.GetId(),
		TenantId:     item.GetTenantId(),
		SymbolId:     item.GetSymbolId(),
		MakerFeeRate: item.GetMakerFeeRate(),
		TakerFeeRate: item.GetTakerFeeRate(),
		BuyEnabled:   item.GetBuyEnabled(),
		SellEnabled:  item.GetSellEnabled(),
		CreateTimes:  item.GetCreateTimes(),
		UpdateTimes:  item.GetUpdateTimes(),
	}
}

func tradeSymbolContractToTypes(item *trade.TradeSymbolContract) types.TradeSymbolContract {
	return types.TradeSymbolContract{
		Id:                     item.GetId(),
		TenantId:               item.GetTenantId(),
		SymbolId:               item.GetSymbolId(),
		ContractSize:           item.GetContractSize(),
		Multiplier:             item.GetMultiplier(),
		MaintenanceMarginRate:  item.GetMaintenanceMarginRate(),
		InitialMarginRate:      item.GetInitialMarginRate(),
		MakerFeeRate:           item.GetMakerFeeRate(),
		TakerFeeRate:           item.GetTakerFeeRate(),
		FundingIntervalMinutes: item.GetFundingIntervalMinutes(),
		DeliveryTime:           item.GetDeliveryTime(),
		SupportCross:           item.GetSupportCross(),
		SupportIsolated:        item.GetSupportIsolated(),
		BuyEnabled:             item.GetBuyEnabled(),
		SellEnabled:            item.GetSellEnabled(),
		CreateTimes:            item.GetCreateTimes(),
		UpdateTimes:            item.GetUpdateTimes(),
	}
}
