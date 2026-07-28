package adminlogic

import (
	"context"
	"errors"
	"strings"

	"wklive/common/generate"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/logic/helpers"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type IgnoreContractReconciliationIssueLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIgnoreContractReconciliationIssueLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IgnoreContractReconciliationIssueLogic {
	return &IgnoreContractReconciliationIssueLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *IgnoreContractReconciliationIssueLogic) IgnoreContractReconciliationIssue(in *trade.IgnoreContractReconciliationIssueReq) (*trade.CommonResp, error) {
	tenantID := helpers.AdminTenantID(l.ctx, in.GetTenantId())
	operatorID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil || operatorID <= 0 {
		return &trade.CommonResp{Base: helper.ErrResp(i18n.OperationNotAllowed, "missing admin operator identity")}, nil
	}
	reason := strings.TrimSpace(in.GetReason())
	if reason == "" || len(reason) > 500 {
		return &trade.CommonResp{Base: helper.ErrResp(i18n.OperationNotAllowed, "ignore reason is required and must not exceed 500 bytes")}, nil
	}
	eventNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "order_id", "TRE", "")
	if err != nil {
		return nil, err
	}
	notFound, invalidStatus := false, false
	now := utils.NowMillis()
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		issueModel := models.NewTContractReconciliationIssueModel(conn, l.svcCtx.Config.CacheRedis)
		issue, findErr := issueModel.FindOneForUpdate(ctx, in.GetId())
		if errors.Is(findErr, models.ErrNotFound) || (findErr == nil && issue.TenantId != tenantID) {
			notFound = true
			return nil
		}
		if findErr != nil {
			return findErr
		}
		if issue.Status != int64(trade.ContractReconciliationIssueStatus_CONTRACT_RECONCILIATION_ISSUE_STATUS_OPEN) {
			invalidStatus = true
			return nil
		}
		issue.Status = int64(trade.ContractReconciliationIssueStatus_CONTRACT_RECONCILIATION_ISSUE_STATUS_IGNORED)
		issue.OperatorId = operatorID
		issue.ResolutionReason = reason
		issue.ResolvedAt = now
		issue.UpdateTimes = now
		if updateErr := issueModel.Update(ctx, issue); updateErr != nil {
			return updateErr
		}
		_, insertErr := models.NewTBizTradeEventModel(conn, l.svcCtx.Config.CacheRedis).Insert(ctx, &models.TBizTradeEvent{
			TenantId:      tenantID,
			EventNo:       eventNo,
			EventType:     "CONTRACT_RECONCILIATION_ISSUE_IGNORED",
			BizId:         issue.IssueKey,
			BizType:       "contract_reconciliation_issue",
			OperatorId:    operatorID,
			Source:        int64(trade.SourceType_SOURCE_TYPE_ADMIN),
			EventStatus:   int64(trade.EventStatus_EVENT_STATUS_PENDING),
			MaxRetryCount: 20,
			NextRetryAt:   now,
			Payload:       helpers.NormalizeTradeEventJSON(reason),
			CreateTimes:   now,
			UpdateTimes:   now,
		})
		return insertErr
	})
	if err != nil {
		return nil, err
	}
	if notFound {
		return &trade.CommonResp{Base: helper.ErrResp(i18n.BusinessDataNotFound, "reconciliation issue not found")}, nil
	}
	if invalidStatus {
		return &trade.CommonResp{Base: helper.ErrResp(i18n.OperationNotAllowed, "only open reconciliation issues can be ignored")}, nil
	}
	return &trade.CommonResp{Base: helper.OkResp()}, nil
}
