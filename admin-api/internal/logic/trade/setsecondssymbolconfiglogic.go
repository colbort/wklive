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

type SetSecondsSymbolConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetSecondsSymbolConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetSecondsSymbolConfigLogic {
	return &SetSecondsSymbolConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetSecondsSymbolConfigLogic) SetSecondsSymbolConfig(req *types.SetSecondsSymbolConfigReq) (resp *types.CommonResp, err error) {
	return logicutil.Proxy[types.CommonResp](l.ctx, req, l.svcCtx.TradeCli.SetSecondsSymbolConfig)
}
