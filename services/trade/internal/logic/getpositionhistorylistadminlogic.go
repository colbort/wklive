package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPositionHistoryListAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPositionHistoryListAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPositionHistoryListAdminLogic {
	return &GetPositionHistoryListAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取持仓历史列表
func (l *GetPositionHistoryListAdminLogic) GetPositionHistoryListAdmin(in *trade.GetPositionHistoryListAdminReq) (*trade.GetPositionHistoryListAdminResp, error) {
	// todo: add your logic here and delete this line

	return &trade.GetPositionHistoryListAdminResp{}, nil
}
