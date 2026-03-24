package logic

import (
	"context"

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

	prevCursor := in.Page.Cursor
	if prevCursor < 0 {
		prevCursor = 0
	}
	nextCursor := int64(0)
	if int64(len(items)) == in.Page.Limit {
		lastItem := items[len(items)-1]
		nextCursor = lastItem.Id
	}
	hasPrev := prevCursor > 0
	hasNext := int64(len(items)) == in.Page.Limit

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
			StartTime:      item.StartTime.Time.Unix(),
			EndTime:        item.EndTime.Time.Unix(),
			CreateTime:     item.CreateTime.Unix(),
		})
	}

	return &system.SysCronJobLogListResp{
		Base: &system.RespBase{
			Code:       0,
			Msg:        "success",
			Total:      total,
			HasPrev:    hasPrev,
			HasNext:    hasNext,
			NextCursor: nextCursor,
			PrevCursor: prevCursor,
		},
		Data: data,
	}, nil
}
