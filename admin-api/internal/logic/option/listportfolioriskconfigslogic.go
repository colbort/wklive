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

type ListPortfolioRiskConfigsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListPortfolioRiskConfigsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPortfolioRiskConfigsLogic {
	return &ListPortfolioRiskConfigsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListPortfolioRiskConfigsLogic) ListPortfolioRiskConfigs(req *types.ListPortfolioRiskConfigsReq) (resp *types.ListPortfolioRiskConfigsResp, err error) {
	return logicutil.Proxy[types.ListPortfolioRiskConfigsResp](
		l.ctx, req, l.svcCtx.OptionCli.ListPortfolioRiskConfigs,
	)
}
