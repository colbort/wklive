package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppGetPositionDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppGetPositionDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppGetPositionDetailLogic {
	return &AppGetPositionDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个持仓详情
func (l *AppGetPositionDetailLogic) AppGetPositionDetail(in *option.AppGetPositionDetailReq) (*option.AppGetPositionDetailResp, error) {
	// todo: add your logic here and delete this line

	return &option.AppGetPositionDetailResp{}, nil
}
