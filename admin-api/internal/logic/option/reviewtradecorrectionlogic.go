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

type ReviewTradeCorrectionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReviewTradeCorrectionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewTradeCorrectionLogic {
	return &ReviewTradeCorrectionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReviewTradeCorrectionLogic) ReviewTradeCorrection(req *types.ReviewTradeCorrectionReq) (resp *types.GetTradeCorrectionResp, err error) {
	return logicutil.Proxy[types.GetTradeCorrectionResp](
		l.ctx, req, l.svcCtx.OptionCli.ReviewTradeCorrection,
	)
}
