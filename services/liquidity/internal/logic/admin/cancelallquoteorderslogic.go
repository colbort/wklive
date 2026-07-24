package adminlogic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"wklive/common/helper"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelAllQuoteOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelAllQuoteOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelAllQuoteOrdersLogic {
	return &CancelAllQuoteOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CancelAllQuoteOrdersLogic) CancelAllQuoteOrders(in *liquidity.SymbolActionReq) (*liquidity.CommonResp, error) {
	config, err := l.svcCtx.SymbolConfigModel.FindOne(l.ctx, in.ConfigId)
	if err != nil {
		return nil, err
	}
	if config.TenantId != in.TenantId {
		return nil, fmt.Errorf("symbol config not found")
	}
	if config.Version != in.Version {
		return nil, fmt.Errorf("symbol config version conflict")
	}
	now := time.Now().UnixMilli()
	reason := strings.TrimSpace(in.Reason)
	if reason == "" {
		reason = "manual cancel all"
	}
	err = l.svcCtx.QuoteOrderModel.CancelActiveByConfig(
		l.ctx, in.TenantId, in.ConfigId, reason, now,
		int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_PENDING_SUBMIT),
		int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_CANCELED),
		int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_CANCELING),
	)
	if err != nil {
		return nil, err
	}
	return &liquidity.CommonResp{Base: helper.OkResp()}, nil
}
