package logic

import (
	"context"

	"wklive/common/pageutil"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysCronJobLogListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysCronJobLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysCronJobLogListLogic {
	return &SysCronJobLogListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 系统定时任务日志列表
func (l *SysCronJobLogListLogic) SysCronJobLogList(in *system.SysCronJobLogListReq) (*system.SysCronJobLogListResp, error) {
	items, total, err := l.svcCtx.JobLogModel.FindPage(l.ctx, in.Page.Cursor, in.Page.Limit, in.JobId, in.JobName, in.InvokeTarget, in.Status)
	if err != nil {
		return nil, err
	}

	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}

	data := make([]*system.SysCronJobLogItem, 0, len(items))
	for _, item := range items {
		data = append(data, &system.SysCronJobLogItem{
			Id:             item.Id,
			JobId:          item.JobId,
			JobName:        item.JobName,
			InvokeTarget:   item.InvokeTarget,
			CronExpression: item.CronExpression.String,
			Status:         item.Status,
			Message:        item.Message.String,
			ExceptionInfo:  item.ExceptionInfo.String,
			StartTime:      item.StartTime,
			EndTime:        item.EndTime,
			CreateTimes:    item.CreateTimes,
		})
	}

	return &system.SysCronJobLogListResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastID),
		Data: data,
	}, nil
}
