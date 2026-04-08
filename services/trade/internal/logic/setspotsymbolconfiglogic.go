package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetSpotSymbolConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetSpotSymbolConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetSpotSymbolConfigLogic {
	return &SetSpotSymbolConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置现货交易对配置
func (l *SetSpotSymbolConfigLogic) SetSpotSymbolConfig(in *trade.SetSpotSymbolConfigReq) (*trade.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &trade.AdminCommonResp{}, nil
}
