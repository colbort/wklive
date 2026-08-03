package adminlogic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserTradeControlAuditsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUserTradeControlAuditsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserTradeControlAuditsLogic {
	return &ListUserTradeControlAuditsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页获取用户交易控制审计
func (l *ListUserTradeControlAuditsLogic) ListUserTradeControlAudits(in *trade.ListUserTradeControlAuditsReq) (*trade.ListUserTradeControlAuditsResp, error) {
	if in.ScopeType < 0 || in.ScopeType > 2 {
		return nil, errors.New("invalid control scope type")
	}
	tenantID, allowed, forbidden, err := utils.ResolveAdminTenantReadScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &trade.ListUserTradeControlAuditsResp{Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx))}, nil
	}
	cursor, limit := pageutil.Input(in.Page)
	rows, total, err := l.svcCtx.TradeUserControlAuditModel.FindPage(l.ctx, models.UserTradeControlFilter{
		TenantId: tenantID, UserId: in.UserId, ControlId: in.ControlId, ScopeType: in.ScopeType,
	}, cursor, limit)
	if err != nil {
		return nil, err
	}
	resp := &trade.ListUserTradeControlAuditsResp{}
	last := int64(0)
	for _, row := range rows {
		resp.Data = append(resp.Data, &trade.TradeUserControlAudit{
			Id: row.Id, TenantId: row.TenantId, ControlId: row.ControlId, ScopeType: row.ScopeType,
			UserId: row.UserId, ChangeType: row.ChangeType, BeforeJson: row.BeforeJson.String,
			AfterJson: row.AfterJson.String, OperatorId: row.OperatorId, Source: trade.SourceType(row.Source),
			Reason: row.Reason, RequestId: row.RequestId, CreateTimes: row.CreateTimes,
		})
		last = row.Id
	}
	resp.Base = pageutil.Base(cursor, limit, len(rows), total, last)
	return resp, nil
}
