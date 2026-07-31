// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package option

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListReconciliationIssuesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListReconciliationIssuesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListReconciliationIssuesLogic {
	return &ListReconciliationIssuesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListReconciliationIssuesLogic) ListReconciliationIssues(req *types.ListReconciliationIssuesReq) (resp *types.ListReconciliationIssuesResp, err error) {
	return logicutil.Proxy[types.ListReconciliationIssuesResp](
		l.ctx, req, l.svcCtx.OptionCli.ListReconciliationIssues,
	)
}
