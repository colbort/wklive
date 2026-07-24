// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package liquidity

import (
	"context"
	"strings"

	"wklive/liquidity-admin-api/internal/logicutil"
	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"
	pb "wklive/proto/liquidity"

	"github.com/zeromicro/go-zero/core/logx"
)

type SymbolActionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSymbolActionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SymbolActionLogic {
	return &SymbolActionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SymbolActionLogic) SymbolAction(req *types.SymbolActionReq) (resp *types.RespBase, err error) {
	tenantID, userID, err := logicutil.Identity(l.ctx)
	if err != nil {
		return nil, err
	}
	in := &pb.SymbolActionReq{
		TenantId: tenantID, ConfigId: req.Id, Version: req.Version, OperatorId: userID, Reason: req.Reason,
	}
	var out *pb.CommonResp
	switch strings.ToLower(req.Action) {
	case "start":
		out, err = l.svcCtx.LiquidityCli.StartSymbolLiquidity(l.ctx, in)
	case "pause":
		out, err = l.svcCtx.LiquidityCli.PauseSymbolLiquidity(l.ctx, in)
	case "stop":
		out, err = l.svcCtx.LiquidityCli.StopSymbolLiquidity(l.ctx, in)
	default:
		return nil, invalidAction(req.Action)
	}
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.RespBase](out.Base), nil
}
