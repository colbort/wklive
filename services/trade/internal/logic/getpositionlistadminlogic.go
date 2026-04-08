package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPositionListAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPositionListAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPositionListAdminLogic {
	return &GetPositionListAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取后台持仓列表
func (l *GetPositionListAdminLogic) GetPositionListAdmin(in *trade.GetPositionListAdminReq) (*trade.GetPositionListAdminResp, error) {
	// todo: add your logic here and delete this line

	return &trade.GetPositionListAdminResp{}, nil
}
