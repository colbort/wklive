// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package option

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTradeCorrectionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateTradeCorrectionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTradeCorrectionLogic {
	return &CreateTradeCorrectionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTradeCorrectionLogic) CreateTradeCorrection(req *types.CreateTradeCorrectionReq) (resp *types.GetTradeCorrectionResp, err error) {
	return logicutil.Proxy[types.GetTradeCorrectionResp](
		l.ctx, req, l.svcCtx.OptionCli.CreateTradeCorrection,
	)
}
