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

type ListAssetInstructionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListAssetInstructionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAssetInstructionsLogic {
	return &ListAssetInstructionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAssetInstructionsLogic) ListAssetInstructions(req *types.ListAssetInstructionsReq) (resp *types.ListAssetInstructionsResp, err error) {
	return logicutil.Proxy[types.ListAssetInstructionsResp](
		l.ctx, req, l.svcCtx.OptionCli.ListAssetInstructions,
	)
}
