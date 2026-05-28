package logic

import (
	"context"
	"errors"

	"wklive/common/pageutil"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPositionHistoryListAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPositionHistoryListAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPositionHistoryListAdminLogic {
	return &GetPositionHistoryListAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取持仓历史列表
func (l *GetPositionHistoryListAdminLogic) GetPositionHistoryListAdmin(in *trade.GetPositionHistoryListAdminReq) (*trade.GetPositionHistoryListAdminResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	data, total, err := l.svcCtx.ContractPositionHistModel.FindPage(l.ctx, models.ContractPositionHistoryPageFilter{
		TenantId:   in.TenantId,
		UserId:     in.UserId,
		SymbolId:   in.SymbolId,
		MarketType: int64(in.MarketType),
		PositionId: int64(in.PositionId),
		ActionType: int64(in.ActionType),
		TimeStart:  in.TimeRange.StartTime,
		TimeEnd:    in.TimeRange.EndTime,
	}, cursor, limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(data) > 0 {
		lastID = int64(data[len(data)-1].Id)
	}
	resp := &trade.GetPositionHistoryListAdminResp{Base: pageutil.Base(cursor, limit, len(data), total, lastID)}
	for _, item := range data {
		resp.Data = append(resp.Data, positionHistoryToProto(item))
	}
	return resp, nil
}
