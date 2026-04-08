package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMarginAccountListAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMarginAccountListAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMarginAccountListAdminLogic {
	return &GetMarginAccountListAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取保证金账户列表
func (l *GetMarginAccountListAdminLogic) GetMarginAccountListAdmin(in *trade.GetMarginAccountListAdminReq) (*trade.GetMarginAccountListAdminResp, error) {
	// todo: add your logic here and delete this line

	return &trade.GetMarginAccountListAdminResp{}, nil
}
