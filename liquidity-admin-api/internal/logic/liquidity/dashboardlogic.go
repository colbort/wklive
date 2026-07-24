// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package liquidity

import (
	"context"

	"wklive/liquidity-admin-api/internal/logicutil"
	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"
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
	tenantID, err := logicutil.TenantID(l.ctx)
	if err != nil {
		return nil, err
	}
	const limit = 1
	symbols, err := l.svcCtx.LiquidityCli.GetSymbolConfigList(l.ctx, &pb.GetSymbolConfigListReq{
		TenantId: tenantID, Status: pb.SymbolLiquidityStatus_SYMBOL_LIQUIDITY_STATUS_RUNNING, Limit: limit,
	})
	if err != nil {
		return nil, err
	}
	providers, err := l.svcCtx.LiquidityCli.GetProviderList(l.ctx, &pb.GetProviderListReq{
		TenantId: tenantID, Status: pb.ProviderStatus_PROVIDER_STATUS_ENABLED, Limit: limit,
	})
	if err != nil {
		return nil, err
	}
	quotes, err := l.svcCtx.LiquidityCli.GetQuoteOrderList(l.ctx, &pb.GetQuoteOrderListReq{
		TenantId: tenantID, Status: pb.QuoteOrderStatus_QUOTE_ORDER_STATUS_OPEN, Limit: limit,
	})
	if err != nil {
		return nil, err
	}
	risks, err := l.svcCtx.LiquidityCli.GetRiskEventList(l.ctx, &pb.GetRiskEventListReq{
		TenantId: tenantID, Status: pb.RiskEventStatus_RISK_EVENT_STATUS_PENDING, Limit: limit,
	})
	if err != nil {
		return nil, err
	}
	return &types.DashboardResp{
		RespBase:         types.RespBase{Code: 0, Msg: "success"},
		RunningSymbols:   symbols.GetPage().GetTotal(),
		HealthyProviders: providers.GetPage().GetTotal(),
		OpenQuotes:       quotes.GetPage().GetTotal(),
		PendingRisks:     risks.GetPage().GetTotal(),
	}, nil
}
