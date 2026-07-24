package liquiditylogic

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"wklive/common/helper"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"
	"wklive/services/liquidity/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReportTradeFillLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReportTradeFillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReportTradeFillLogic {
	return &ReportTradeFillLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ReportTradeFillLogic) ReportTradeFill(in *liquidity.ReportTradeFillReq) (*liquidity.CommonResp, error) {
	if in.SymbolId <= 0 || strings.TrimSpace(in.EventNo) == "" {
		return nil, fmt.Errorf("symbol_id and event_no are required")
	}
	if existing, err := l.svcCtx.EventInboxModel.FindOneByConsumerEventNo(l.ctx, "liquidity.trade_fill", in.EventNo); err == nil {
		if existing.Status == 2 {
			return &liquidity.CommonResp{Base: helper.OkResp()}, nil
		}
	} else if err != models.ErrNotFound {
		return nil, err
	}
	for name, value := range map[string]string{"price": in.Price, "qty": in.Qty, "amount": in.Amount} {
		number, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil || number < 0 {
			return nil, fmt.Errorf("%s is invalid", name)
		}
	}
	payload, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	now := time.Now().UnixMilli()
	row := &models.TLiquidityEventInbox{
		Consumer: "liquidity.trade_fill", EventNo: strings.TrimSpace(in.EventNo),
		EventType: "TRADE_FILL", AggregateType: "TRADE_ORDER", AggregateId: strconv.FormatInt(in.OrderId, 10),
		Payload: sql.NullString{String: string(payload), Valid: true}, Status: 2,
		ProcessedAt: now, CreateTimes: now, UpdateTimes: now,
	}
	if _, err := l.svcCtx.EventInboxModel.Insert(l.ctx, row); err != nil {
		if existing, findErr := l.svcCtx.EventInboxModel.FindOneByConsumerEventNo(l.ctx, "liquidity.trade_fill", in.EventNo); findErr == nil && existing.Status == 2 {
			return &liquidity.CommonResp{Base: helper.OkResp()}, nil
		}
		return nil, err
	}
	return &liquidity.CommonResp{Base: helper.OkResp()}, nil
}
