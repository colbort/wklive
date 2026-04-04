package logic

import (
	"context"

	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysCronJobListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysCronJobListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysCronJobListLogic {
	return &SysCronJobListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 系统定时任务列表
func (l *SysCronJobListLogic) SysCronJobList(in *system.SysCronJobListReq) (*system.SysCronJobListResp, error) {
	items, total, err := l.svcCtx.JobModel.FindPage(l.ctx, in.Page.Cursor, in.Page.Limit, in.Keyword, in.JobName, in.JobGroup, in.Status)
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

	data := make([]*system.SysCronJobItem, 0, len(items))
	for _, item := range items {
		data = append(data, &system.SysCronJobItem{
			Id:             item.Id,
			JobName:        item.JobName,
			JobGroup:       item.JobGroup,
			InvokeTarget:   item.InvokeTarget,
			CronExpression: item.CronExpression,
			Status:         item.Status,
			Remark:         item.Remark.String,
			CreateBy:       item.CreateBy.String,
			CreateTimes:    item.CreateTimes,
			UpdateBy:       item.UpdateBy.String,
			UpdateTimes:    item.UpdateTimes,
		})
	}

	return &system.SysCronJobListResp{
		Base: &system.RespBase{
			Code:       200,
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
