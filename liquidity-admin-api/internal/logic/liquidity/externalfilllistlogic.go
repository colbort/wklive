// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package liquidity

import (
	"context"

	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExternalFillListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewExternalFillListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExternalFillListLogic {
	return &ExternalFillListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ExternalFillListLogic) ExternalFillList(req *types.ExternalFillQuery) (resp *types.ExternalFillListResp, err error) {
	return externalFillList(l.ctx, l.svcCtx, req)
}
