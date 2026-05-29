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

type GetSymbolLeverageConfigListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSymbolLeverageConfigListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSymbolLeverageConfigListLogic {
	return &GetSymbolLeverageConfigListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取交易对杠杆档位配置列表
func (l *GetSymbolLeverageConfigListLogic) GetSymbolLeverageConfigList(in *trade.GetSymbolLeverageConfigListReq) (*trade.GetSymbolLeverageConfigListResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	data, total, err := l.svcCtx.SymbolLeverageCfgModel.FindPage(l.ctx, in.TenantId, in.SymbolId, int64(in.MarketType), int64(in.MarginMode), in.Status, cursor, limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(data) > 0 {
		lastID = data[len(data)-1].Id
	}
	resp := &trade.GetSymbolLeverageConfigListResp{Base: pageutil.Base(cursor, limit, len(data), total, lastID)}
	for _, item := range data {
		resp.Data = append(resp.Data, symbolLeverageConfigToProto(item))
	}
	return resp, nil
}
