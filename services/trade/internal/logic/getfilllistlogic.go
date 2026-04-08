package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFillListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFillListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFillListLogic {
	return &GetFillListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取成交记录列表
func (l *GetFillListLogic) GetFillList(in *trade.GetFillListReq) (*trade.GetFillListResp, error) {
	// todo: add your logic here and delete this line

	return &trade.GetFillListResp{}, nil
}
