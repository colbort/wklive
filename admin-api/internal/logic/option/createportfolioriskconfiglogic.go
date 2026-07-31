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

type CreatePortfolioRiskConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreatePortfolioRiskConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePortfolioRiskConfigLogic {
	return &CreatePortfolioRiskConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreatePortfolioRiskConfigLogic) CreatePortfolioRiskConfig(req *types.CreatePortfolioRiskConfigReq) (resp *types.GetPortfolioRiskConfigResp, err error) {
	return logicutil.Proxy[types.GetPortfolioRiskConfigResp](
		l.ctx, req, l.svcCtx.OptionCli.CreatePortfolioRiskConfig,
	)
}
