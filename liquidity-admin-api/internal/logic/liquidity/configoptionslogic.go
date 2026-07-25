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

type ConfigOptionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConfigOptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfigOptionsLogic {
	return &ConfigOptionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ConfigOptionsLogic) ConfigOptions() (resp *types.ConfigOptionsResp, err error) {
	out, err := l.svcCtx.LiquidityCli.GetConfigOptions(l.ctx, &pb.GetConfigOptionsReq{})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.ConfigOptionsResp](out), nil
}
