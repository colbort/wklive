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

type SetUserTradeConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetUserTradeConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetUserTradeConfigLogic {
	return &SetUserTradeConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetUserTradeConfigLogic) SetUserTradeConfig(req *types.SetUserTradeConfigReq) (resp *types.AdminCommonResp, err error) {
	return logicutil.Proxy[types.AdminCommonResp](l.ctx, req, l.svcCtx.TradeCli.SetUserTradeConfig)
}
