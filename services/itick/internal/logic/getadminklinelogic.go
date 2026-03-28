package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAdminKlineLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAdminKlineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAdminKlineLogic {
	return &GetAdminKlineLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// K线查看
func (l *GetAdminKlineLogic) GetAdminKline(in *itick.GetAdminKlineReq) (*itick.GetAdminKlineResp, error) {
	// todo: add your logic here and delete this line

	return &itick.GetAdminKlineResp{}, nil
}
