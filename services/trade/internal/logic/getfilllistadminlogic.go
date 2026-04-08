package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFillListAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFillListAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFillListAdminLogic {
	return &GetFillListAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetFillListAdminLogic) GetFillListAdmin(in *trade.GetFillListAdminReq) (*trade.GetFillListAdminResp, error) {
	// todo: add your logic here and delete this line

	return &trade.GetFillListAdminResp{}, nil
}
