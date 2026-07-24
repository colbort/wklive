// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package liquidity

import (
	"context"

	"wklive/liquidity-admin-api/internal/logicutil"
	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"
	pb "wklive/proto/liquidity"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProviderCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProviderCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProviderCreateLogic {
	return &ProviderCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProviderCreateLogic) ProviderCreate(req *types.CreateProviderReq) (resp *types.RespBase, err error) {
	tenantID, userID, err := logicutil.Identity(l.ctx)
	if err != nil {
		return nil, err
	}
	in := logicutil.Convert[pb.CreateProviderReq](req)
	in.TenantId, in.OperatorId = tenantID, userID
	out, err := l.svcCtx.LiquidityCli.CreateProvider(l.ctx, in)
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.RespBase](out.Base), nil
}
