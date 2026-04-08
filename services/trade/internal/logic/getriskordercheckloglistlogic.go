package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRiskOrderCheckLogListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRiskOrderCheckLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRiskOrderCheckLogListLogic {
	return &GetRiskOrderCheckLogListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetRiskOrderCheckLogListLogic) GetRiskOrderCheckLogList(in *trade.GetRiskOrderCheckLogListReq) (*trade.GetRiskOrderCheckLogListResp, error) {
	// todo: add your logic here and delete this line

	return &trade.GetRiskOrderCheckLogListResp{}, nil
}
