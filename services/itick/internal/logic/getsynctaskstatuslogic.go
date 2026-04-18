package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/i18n"
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
	item, err := l.svcCtx.ItickSyncTaskModel.FindOneByTaskNo(l.ctx, in.TaskNo)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return &itick.GetSyncTaskStatusResp{
			Base: helper.GetErrResp(1, i18n.Translate(i18n.BusinessDataNotFound, l.ctx)),
		}, nil
	}

	return &itick.GetSyncTaskStatusResp{
		Base:    helper.OkResp(),
		TaskNo:  item.TaskNo,
		Status:  int32(item.Status),
		Message: item.Message,
	}, nil
}
