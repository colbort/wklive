// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/itick"

	"github.com/zeromicro/go-zero/core/logx"
)

type InitTenantItickDisplayLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInitTenantItickDisplayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitTenantItickDisplayLogic {
	return &InitTenantItickDisplayLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InitTenantItickDisplayLogic) InitTenantItickDisplay(req *types.InitTenantItickDisplayReq) (resp *types.InitTenantItickDisplayResp, err error) {
	result, err := l.svcCtx.ItickCli.InitTenantItickDisplay(l.ctx, &itick.InitTenantItickDisplayReq{
		TenantId:  req.TenantId,
		Overwrite: req.Overwrite,
	})
	if err != nil {
		return nil, err
	}
	return &types.InitTenantItickDisplayResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		CategoryCount: result.CategoryCount,
		ProductCount:  result.ProductCount,
	}, nil
}
