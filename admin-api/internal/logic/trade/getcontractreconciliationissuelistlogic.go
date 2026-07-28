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

type GetContractReconciliationIssueListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetContractReconciliationIssueListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetContractReconciliationIssueListLogic {
	return &GetContractReconciliationIssueListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetContractReconciliationIssueListLogic) GetContractReconciliationIssueList(req *types.GetContractReconciliationIssueListReq) (resp *types.GetContractReconciliationIssueListResp, err error) {
	return logicutil.Proxy[types.GetContractReconciliationIssueListResp](l.ctx, req, l.svcCtx.TradeCli.GetContractReconciliationIssueList)
}
