package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type InitTenantItickDisplayLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInitTenantItickDisplayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitTenantItickDisplayLogic {
	return &InitTenantItickDisplayLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 初始化租户展示配置
func (l *InitTenantItickDisplayLogic) InitTenantItickDisplay(in *itick.InitTenantItickDisplayReq) (*itick.InitTenantItickDisplayResp, error) {
	// todo: add your logic here and delete this line

	return &itick.InitTenantItickDisplayResp{}, nil
}
