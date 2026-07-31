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

type ResumeContractTradingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResumeContractTradingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResumeContractTradingLogic {
	return &ResumeContractTradingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResumeContractTradingLogic) ResumeContractTrading(req *types.ResumeContractTradingReq) (resp *types.GetTradingHaltResp, err error) {
	return logicutil.Proxy[types.GetTradingHaltResp](
		l.ctx, req, l.svcCtx.OptionCli.ResumeContractTrading,
	)
}
