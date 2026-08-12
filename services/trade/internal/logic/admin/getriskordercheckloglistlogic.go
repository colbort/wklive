package adminlogic

import (
	"context"
	"errors"
	"wklive/services/trade/internal/logic/helpers"

	"wklive/common/pageutil"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRiskOrderCheckLogListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRiskOrderCheckLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRiskOrderCheckLogListLogic {
	return &GetRiskOrderCheckLogListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取风控订单校验日志列表
func (l *GetRiskOrderCheckLogListLogic) GetRiskOrderCheckLogList(in *trade.GetRiskOrderCheckLogListReq) (*trade.GetRiskOrderCheckLogListResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	data, total, err := l.svcCtx.RiskOrderCheckLogModel.FindPage(l.ctx, models.RiskOrderCheckLogPageFilter{
		TenantId:    in.TenantId,
		UserId:      in.UserId,
		SymbolId:    in.SymbolId,
		ProductType: int64(in.ProductType),
		CheckType:   int64(in.CheckType),
		CheckResult: int64(in.CheckResult),
		TimeStart:   in.TimeRange.StartTime,
		TimeEnd:     in.TimeRange.EndTime,
	}, cursor, limit, pageutil.Count(in.Page))
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	lastID := int64(0)
	if len(data) > 0 {
		lastID = int64(data[len(data)-1].Id)
	}
	resp := &trade.GetRiskOrderCheckLogListResp{Base: pageutil.Base(cursor, limit, len(data), total, lastID)}
	for _, item := range data {
		resp.Data = append(resp.Data, helpers.RiskOrderCheckLogToProto(item))
	}
	return resp, nil
}
