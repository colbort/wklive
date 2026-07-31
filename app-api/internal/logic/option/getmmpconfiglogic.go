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

type GetMMPConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 查询当前用户指定做市报价组的 MMP 状态
func NewGetMMPConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMMPConfigLogic {
	return &GetMMPConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMMPConfigLogic) GetMMPConfig(req *types.GetMMPConfigReq) (resp *types.GetMMPConfigResp, err error) {
	return logicutil.Proxy[types.GetMMPConfigResp](
		l.ctx, req, l.svcCtx.OptionCli.GetMMPConfig,
	)
}
