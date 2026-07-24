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

type SymbolConfigCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSymbolConfigCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SymbolConfigCreateLogic {
	return &SymbolConfigCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SymbolConfigCreateLogic) SymbolConfigCreate(req *types.SaveSymbolConfigReq) (resp *types.RespBase, err error) {
	_, userID, err := logicutil.Identity(l.ctx)
	if err != nil {
		return nil, err
	}
	in := logicutil.Convert[pb.SaveSymbolConfigReq](req)
	in.OperatorId = userID
	out, err := l.svcCtx.LiquidityCli.CreateSymbolConfig(l.ctx, in)
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.RespBase](out.Base), nil
}
