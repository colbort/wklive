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

type ListTradeCorrectionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListTradeCorrectionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTradeCorrectionsLogic {
	return &ListTradeCorrectionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTradeCorrectionsLogic) ListTradeCorrections(req *types.ListTradeCorrectionsReq) (resp *types.ListTradeCorrectionsResp, err error) {
	return logicutil.Proxy[types.ListTradeCorrectionsResp](
		l.ctx, req, l.svcCtx.OptionCli.ListTradeCorrections,
	)
}
