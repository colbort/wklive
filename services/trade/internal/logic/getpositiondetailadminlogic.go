package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPositionDetailAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPositionDetailAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPositionDetailAdminLogic {
	return &GetPositionDetailAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取持仓详情
func (l *GetPositionDetailAdminLogic) GetPositionDetailAdmin(in *trade.GetPositionDetailAdminReq) (*trade.GetPositionDetailAdminResp, error) {
	// todo: add your logic here and delete this line

	return &trade.GetPositionDetailAdminResp{}, nil
}
