package adminlogic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/market"
	"wklive/services/market/internal/svc"

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
func (l *GetSyncTaskStatusLogic) GetSyncTaskStatus(in *market.GetSyncTaskStatusReq) (*market.GetSyncTaskStatusResp, error) {
	item, err := l.svcCtx.MarketSyncTaskModel.FindOneByTaskNo(l.ctx, in.TaskNo)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return &market.GetSyncTaskStatusResp{
			Base: helper.ErrResp(i18n.BusinessDataNotFound, i18n.Translate(i18n.BusinessDataNotFound, l.ctx)),
		}, nil
	}

	return &market.GetSyncTaskStatusResp{
		Base: helper.OkResp(),
		Data: &market.GetSyncTaskStatusData{
			TaskNo:  item.TaskNo,
			Status:  int32(item.Status),
			Message: item.Message,
		},
	}, nil
}
