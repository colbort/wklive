package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetUserSymbolLimitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetUserSymbolLimitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetUserSymbolLimitLogic {
	return &SetUserSymbolLimitLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置用户交易对限制
func (l *SetUserSymbolLimitLogic) SetUserSymbolLimit(in *trade.SetUserSymbolLimitReq) (*trade.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &trade.AdminCommonResp{}, nil
}
