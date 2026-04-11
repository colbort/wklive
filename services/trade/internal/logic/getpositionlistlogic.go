package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPositionListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPositionListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPositionListLogic {
	return &GetPositionListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取持仓列表
func (l *GetPositionListLogic) GetPositionList(in *trade.GetPositionListReq) (*trade.GetPositionListResp, error) {
	list, err := l.svcCtx.ContractPositionModel.FindList(l.ctx, models.ContractPositionPageFilter{
		TenantId:   int64(in.TenantId),
		UserId:     int64(in.UserId),
		SymbolId:   int64(in.SymbolId),
		MarketType: int64(in.MarketType),
	})
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	resp := &trade.GetPositionListResp{Base: helper.OkResp()}
	for _, item := range list {
		resp.List = append(resp.List, positionToProto(item))
	}
	return resp, nil
}
