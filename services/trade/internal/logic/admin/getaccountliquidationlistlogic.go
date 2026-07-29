package adminlogic

import (
	"context"

	"wklive/common/pageutil"
	"wklive/proto/trade"
	"wklive/services/trade/internal/logic/helpers"
	"wklive/services/trade/internal/mapper"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAccountLiquidationListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAccountLiquidationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountLiquidationListLogic {
	return &GetAccountLiquidationListLogic{
		ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx),
	}
}

func (l *GetAccountLiquidationListLogic) GetAccountLiquidationList(
	in *trade.GetAccountLiquidationListReq,
) (*trade.GetAccountLiquidationListResp, error) {
	cursor, limit := pageutil.Input(in.GetPage())
	rows, total, err := l.svcCtx.ContractAccountLiqModel.FindPage(l.ctx, models.AdminPageFilter{
		TenantId: helpers.AdminTenantID(l.ctx, in.GetTenantId()),
		UserId:   in.GetUserId(), MarginAsset: in.GetMarginAsset(),
		Status: int64(in.GetStatus()), TimeStart: in.GetTimeRange().GetStartTime(),
		TimeEnd: in.GetTimeRange().GetEndTime(),
	}, cursor, limit)
	if err != nil {
		return nil, err
	}
	last := int64(0)
	resp := &trade.GetAccountLiquidationListResp{}
	for _, row := range rows {
		resp.Data = append(resp.Data, mapper.AccountLiquidationProto(row))
		last = row.Id
	}
	resp.Base = pageutil.Base(cursor, limit, len(rows), total, last)
	return resp, nil
}
