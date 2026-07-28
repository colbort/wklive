// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type IgnoreContractReconciliationIssueLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewIgnoreContractReconciliationIssueLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IgnoreContractReconciliationIssueLogic {
	return &IgnoreContractReconciliationIssueLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *IgnoreContractReconciliationIssueLogic) IgnoreContractReconciliationIssue(req *types.IgnoreContractReconciliationIssueReq) (resp *types.CommonResp, err error) {
	return logicutil.Proxy[types.CommonResp](l.ctx, req, l.svcCtx.TradeCli.IgnoreContractReconciliationIssue)
}
