package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSyncTaskStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSyncTaskStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSyncTaskStatusLogic {
	return &GetSyncTaskStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取同步任务状态
func (l *GetSyncTaskStatusLogic) GetSyncTaskStatus(in *itick.GetSyncTaskStatusReq) (*itick.GetSyncTaskStatusResp, error) {
	// todo: add your logic here and delete this line

	return &itick.GetSyncTaskStatusResp{}, nil
}
