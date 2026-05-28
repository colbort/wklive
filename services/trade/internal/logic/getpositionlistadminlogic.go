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

type GetPositionListAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPositionListAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPositionListAdminLogic {
	return &GetPositionListAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取后台持仓列表
func (l *GetPositionListAdminLogic) GetPositionListAdmin(in *trade.GetPositionListAdminReq) (*trade.GetPositionListAdminResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	data, total, err := l.svcCtx.ContractPositionModel.FindPage(l.ctx, models.ContractPositionPageFilter{
		TenantId:   in.TenantId,
		UserId:     in.UserId,
		SymbolId:   in.SymbolId,
		MarketType: int64(in.MarketType),
	}, cursor, limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(data) > 0 {
		lastID = int64(data[len(data)-1].Id)
	}
	resp := &trade.GetPositionListAdminResp{Base: pageutil.Base(cursor, limit, len(data), total, lastID)}
	for _, item := range data {
		resp.Data = append(resp.Data, positionToProto(item))
	}
	return resp, nil
}
