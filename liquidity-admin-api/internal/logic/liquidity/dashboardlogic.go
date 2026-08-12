// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package liquidity

import (
	"context"

	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"
	"wklive/proto/common"
	pb "wklive/proto/liquidity"

	"github.com/zeromicro/go-zero/core/logx"
)

type DashboardLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDashboardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DashboardLogic {
	return &DashboardLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DashboardLogic) Dashboard() (resp *types.DashboardResp, err error) {
	const limit = 1
	symbols, err := l.svcCtx.LiquidityCli.GetSymbolConfigList(l.ctx, &pb.GetSymbolConfigListReq{
		Status: pb.SymbolLiquidityStatus_SYMBOL_LIQUIDITY_STATUS_RUNNING, Page: &common.PageReq{Limit: limit},
	})
	if err != nil {
		return nil, err
	}
	providers, err := l.svcCtx.LiquidityCli.GetProviderList(l.ctx, &pb.GetProviderListReq{
		Status: pb.ProviderStatus_PROVIDER_STATUS_ENABLED, Page: &common.PageReq{Limit: limit},
	})
	if err != nil {
		return nil, err
	}
	quotes, err := l.svcCtx.LiquidityCli.GetQuoteOrderList(l.ctx, &pb.GetQuoteOrderListReq{
		Status: pb.QuoteOrderStatus_QUOTE_ORDER_STATUS_OPEN, Page: &common.PageReq{Limit: limit},
	})
	if err != nil {
		return nil, err
	}
	risks, err := l.svcCtx.LiquidityCli.GetRiskEventList(l.ctx, &pb.GetRiskEventListReq{
		Status: pb.RiskEventStatus_RISK_EVENT_STATUS_PENDING, Page: &common.PageReq{Limit: limit},
	})
	if err != nil {
		return nil, err
	}
	return &types.DashboardResp{
		RespBase:         types.RespBase{Code: 0, Msg: "success"},
		RunningSymbols:   symbols.GetBase().GetTotal(),
		HealthyProviders: providers.GetBase().GetTotal(),
		OpenQuotes:       quotes.GetBase().GetTotal(),
		PendingRisks:     risks.GetBase().GetTotal(),
	}, nil
}
