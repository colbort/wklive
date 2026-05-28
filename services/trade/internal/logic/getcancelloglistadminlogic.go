package logic

import (
	"context"
	"errors"

	"wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCancelLogListAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCancelLogListAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCancelLogListAdminLogic {
	return &GetCancelLogListAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取撤单日志列表
func (l *GetCancelLogListAdminLogic) GetCancelLogListAdmin(in *trade.GetCancelLogListAdminReq) (*trade.GetCancelLogListAdminResp, error) {
	if in.TenantId <= 0 {
		if tenantId, err := utils.GetTenantIdFromMd(l.ctx); err == nil {
			in.TenantId = tenantId
		}
	}
	cursor, limit := pageutil.Input(in.Page)
	data, total, err := l.svcCtx.TradeCancelLogModel.FindPage(l.ctx, models.TradeCancelLogPageFilter{
		TenantId:     in.TenantId,
		UserId:       in.UserId,
		OrderId:      int64(in.OrderId),
		OrderNo:      in.OrderNo,
		CancelSource: int64(in.CancelSource),
		TimeStart:    in.TimeRange.StartTime,
		TimeEnd:      in.TimeRange.EndTime,
	}, cursor, limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(data) > 0 {
		lastID = int64(data[len(data)-1].Id)
	}
	resp := &trade.GetCancelLogListAdminResp{Base: pageutil.Base(cursor, limit, len(data), total, lastID)}
	for _, item := range data {
		resp.Data = append(resp.Data, cancelLogToProto(item))
	}
	return resp, nil
}
