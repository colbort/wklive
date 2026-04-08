package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCancelLogListAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCancelLogListAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCancelLogListAdminLogic {
	return &GetCancelLogListAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取撤单日志列表
func (l *GetCancelLogListAdminLogic) GetCancelLogListAdmin(in *trade.GetCancelLogListAdminReq) (*trade.GetCancelLogListAdminResp, error) {
	// todo: add your logic here and delete this line

	return &trade.GetCancelLogListAdminResp{}, nil
}
