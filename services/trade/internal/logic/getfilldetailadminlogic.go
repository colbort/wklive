package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFillDetailAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFillDetailAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFillDetailAdminLogic {
	return &GetFillDetailAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取成交详情
func (l *GetFillDetailAdminLogic) GetFillDetailAdmin(in *trade.GetFillDetailAdminReq) (*trade.GetFillDetailAdminResp, error) {
	// todo: add your logic here and delete this line

	return &trade.GetFillDetailAdminResp{}, nil
}
