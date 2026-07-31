// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package option

import (
	"context"

	"wklive/app-api/internal/logicutil"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListOptionChainLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 按标的和精确到期时间查询期权链、24小时成交统计及单边未平仓量
func NewListOptionChainLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOptionChainLogic {
	return &ListOptionChainLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListOptionChainLogic) ListOptionChain(req *types.ListOptionChainReq) (resp *types.ListOptionChainResp, err error) {
	return logicutil.Proxy[types.ListOptionChainResp](l.ctx, req, l.svcCtx.OptionCli.ListOptionChain)
}
