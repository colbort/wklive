package logic

import (
	"context"

	"wklive/common/pageutil"
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
	items, total, err := l.svcCtx.JobModel.FindPage(l.ctx, in.Page.Cursor, in.Page.Limit, in.Keyword, in.JobName, in.JobGroup, jobStatusToModel(in.Status))
	if err != nil {
		return nil, err
	}

	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}

	data := make([]*system.SysCronJobItem, 0, len(items))
	for _, item := range items {
		data = append(data, &system.SysCronJobItem{
			Id:             item.Id,
			JobName:        item.JobName,
			JobGroup:       item.JobGroup,
			InvokeTarget:   item.InvokeTarget,
			CronExpression: item.CronExpression,
			Status:         jobStatusToProto(item.Status),
			Remark:         item.Remark.String,
			CreateBy:       item.CreateBy.String,
			CreateTimes:    item.CreateTimes,
			UpdateBy:       item.UpdateBy.String,
			UpdateTimes:    item.UpdateTimes,
		})
	}

	return &system.SysCronJobListResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastID),
		Data: data,
	}, nil
}
