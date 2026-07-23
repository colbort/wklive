package applogic

import (
	"context"
	"errors"

	pageutil "wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTradesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTradesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTradesLogic {
	return &ListTradesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取成交记录列表
func (l *ListTradesLogic) ListTrades(in *option.UserListTradesReq) (*option.UserListTradesResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	items, total, err := l.svcCtx.OptionTradeModel.FindPage(l.ctx, models.OptionTradePageFilter{
		TenantId:       tenantId,
		ContractId:     in.ContractId,
		UserId:         userId,
		AccountId:      in.AccountId,
		TradeTimeStart: pageutil.TimeRangeStart(in.TradeTimeRange),
		TradeTimeEnd:   pageutil.TimeRangeEnd(in.TradeTimeRange),
	}, cursor, limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	data := make([]*option.OptionTradeDetail, 0, len(items))
	lastID := int64(0)
	for _, item := range items {
		lastID = item.Id
		detail, err := buildTradeDetail(l.ctx, l.svcCtx, item)
		if err != nil {
			return nil, err
		}
		data = append(data, detail)
	}

	return &option.UserListTradesResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID),
		Data: data,
	}, nil
}
