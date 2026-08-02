package adminlogic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListReconciliationIssuesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListReconciliationIssuesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListReconciliationIssuesLogic {
	return &ListReconciliationIssuesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询 Option 与 Asset 对账差异
func (l *ListReconciliationIssuesLogic) ListReconciliationIssues(in *option.ListReconciliationIssuesReq) (*option.ListReconciliationIssuesResp, error) {
	tenantId, allowed, forbidden, err := utils.ResolveAdminTenantReadScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &option.ListReconciliationIssuesResp{
			Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx)),
		}, nil
	}
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.OptionReconciliationIssueModel.FindPage(
		l.ctx, models.OptionReconciliationIssuePageFilter{
			TenantId: tenantId, BizNo: in.BizNo,
			CheckType: int64(in.CheckType), Status: int64(in.Status),
		}, cursor, limit,
	)
	if err != nil {
		return nil, err
	}
	data := make([]*option.OptionReconciliationIssue, 0, len(items))
	lastID := int64(0)
	for _, item := range items {
		lastID = item.Id
		data = append(data, &option.OptionReconciliationIssue{
			Id: item.Id, TenantId: item.TenantId, IssueKey: item.IssueKey,
			CheckType: option.ReconciliationCheckType(item.CheckType),
			BizNo:     item.BizNo, InstructionId: item.InstructionId,
			ExpectedValue: item.ExpectedValue, ActualValue: item.ActualValue,
			Detail: item.Detail, Status: option.ReconciliationIssueStatus(item.Status),
			OccurrenceCount: item.OccurrenceCount, ResolvedAt: item.ResolvedAt,
			CreateTimes: item.CreateTimes, UpdateTimes: item.UpdateTimes,
		})
	}
	return &option.ListReconciliationIssuesResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID),
		Data: data, Total: total,
	}, nil
}
