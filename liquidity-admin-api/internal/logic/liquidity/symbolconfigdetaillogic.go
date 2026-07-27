package liquidity

import (
	"context"

	"wklive/liquidity-admin-api/internal/logicutil"
	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"
	pb "wklive/proto/liquidity"

	"github.com/zeromicro/go-zero/core/logx"
)

type SymbolConfigDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSymbolConfigDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SymbolConfigDetailLogic {
	return &SymbolConfigDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SymbolConfigDetailLogic) SymbolConfigDetail(req *types.SymbolConfigDetailReq) (*types.SymbolConfigDetailResp, error) {
	out, err := l.svcCtx.LiquidityCli.GetSymbolConfigDetail(l.ctx, &pb.GetSymbolConfigDetailReq{Id: req.Id})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.SymbolConfigDetailResp](out), nil
}
