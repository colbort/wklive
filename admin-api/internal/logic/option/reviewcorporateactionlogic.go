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

type ReviewCorporateActionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReviewCorporateActionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewCorporateActionLogic {
	return &ReviewCorporateActionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReviewCorporateActionLogic) ReviewCorporateAction(req *types.ReviewCorporateActionReq) (resp *types.GetCorporateActionResp, err error) {
	return logicutil.Proxy[types.GetCorporateActionResp](
		l.ctx, req, l.svcCtx.OptionCli.ReviewCorporateAction,
	)
}
