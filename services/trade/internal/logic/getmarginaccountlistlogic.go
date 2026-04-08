package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMarginAccountListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMarginAccountListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMarginAccountListLogic {
	return &GetMarginAccountListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取保证金账户列表
func (l *GetMarginAccountListLogic) GetMarginAccountList(in *trade.GetMarginAccountListReq) (*trade.GetMarginAccountListResp, error) {
	// todo: add your logic here and delete this line

	return &trade.GetMarginAccountListResp{}, nil
}
