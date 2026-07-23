// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package itick

import (
	"context"

	"wklive/app-api/internal/logicutil"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/proto/itick"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetKlineIntervalsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetKlineIntervalsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetKlineIntervalsLogic {
	return &GetKlineIntervalsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetKlineIntervalsLogic) GetKlineIntervals() (resp *types.KlineIntervalsResp, err error) {
	return logicutil.Proxy[types.KlineIntervalsResp](l.ctx, &itick.Empty{}, l.svcCtx.ItickCli.GetKlineIntervals)
}
