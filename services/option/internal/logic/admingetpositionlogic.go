package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminGetPositionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminGetPositionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetPositionLogic {
	return &AdminGetPositionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个持仓详情
func (l *AdminGetPositionLogic) AdminGetPosition(in *option.GetPositionReq) (*option.GetPositionResp, error) {
	// todo: add your logic here and delete this line

	return &option.GetPositionResp{}, nil
}
