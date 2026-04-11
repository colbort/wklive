package logic

import (
	"context"
	"database/sql"
	"errors"

	"wklive/common/helper"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTradeEventLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTradeEventLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTradeEventLogic {
	return &CreateTradeEventLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建交易事件
func (l *CreateTradeEventLogic) CreateTradeEvent(in *trade.CreateTradeEventReq) (*trade.InternalCommonResp, error) {
	if in.Event == nil {
		return &trade.InternalCommonResp{Base: helper.OkResp()}, nil
	}
	exists, err := l.svcCtx.BizTradeEventModel.FindOneByTenantIdEventNo(l.ctx, in.Event.TenantId, in.Event.EventNo)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if exists == nil {
		eventNo := in.Event.EventNo
		if eventNo == "" {
			eventNo, err = l.svcCtx.GenerateBizNo(l.ctx, "TRE")
			if err != nil {
				return nil, err
			}
		}
		_, err = l.svcCtx.BizTradeEventModel.Insert(l.ctx, &models.TBizTradeEvent{
			TenantId:      in.Event.TenantId,
			EventNo:       eventNo,
			EventType:     in.Event.EventType,
			BizId:         in.Event.BizId,
			BizType:       in.Event.BizType,
			UserId:        in.Event.UserId,
			SymbolId:      in.Event.SymbolId,
			MarketType:    int64(in.Event.MarketType),
			OperatorId:    in.Event.OperatorId,
			Source:        int64(in.Event.Source),
			EventStatus:   int64(in.Event.EventStatus),
			RetryCount:    int64(in.Event.RetryCount),
			MaxRetryCount: int64(in.Event.MaxRetryCount),
			NextRetryAt:   in.Event.NextRetryAt,
			LastErrorMsg:  in.Event.LastErrorMsg,
			Payload:       in.Event.Payload,
			ExtData:       sql.NullString{String: in.Event.ExtData, Valid: in.Event.ExtData != ""},
			CreateTimes:   in.Event.CreateTimes,
			UpdateTimes:   in.Event.UpdateTimes,
		})
		if err != nil {
			return nil, err
		}
	}
	return &trade.InternalCommonResp{Base: helper.OkResp()}, nil
}
