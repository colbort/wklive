package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateSymbolLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateSymbolLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSymbolLogic {
	return &UpdateSymbolLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新交易对信息
func (l *UpdateSymbolLogic) UpdateSymbol(in *trade.UpdateSymbolReq) (*trade.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &trade.AdminCommonResp{}, nil
}
