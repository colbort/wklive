// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAdminKlineLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAdminKlineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAdminKlineLogic {
	return &GetAdminKlineLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAdminKlineLogic) GetAdminKline(req *types.GetAdminKlineReq) (resp *types.GetAdminKlineResp, err error) {
	// todo: add your logic here and delete this line

	return
}
