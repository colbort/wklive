package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetContractSymbolConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetContractSymbolConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetContractSymbolConfigLogic {
	return &SetContractSymbolConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置合约交易对配置
func (l *SetContractSymbolConfigLogic) SetContractSymbolConfig(in *trade.SetContractSymbolConfigReq) (*trade.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &trade.AdminCommonResp{}, nil
}
