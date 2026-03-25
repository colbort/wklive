// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/system"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysCronJobLogListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysCronJobLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysCronJobLogListLogic {
	return &SysCronJobLogListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysCronJobLogListLogic) SysCronJobLogList(req *types.SysCronJobLogListReq) (resp *types.SysCronJobLogListResp, err error) {
	result, err := l.svcCtx.SystemCli.SysCronJobLogList(l.ctx, &system.SysCronJobLogListReq{
		Page: &system.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
		JobId:        req.JobId,
		JobName:      req.JobName,
		InvokeTarget: req.InvokeTarget,
		Status:       req.Status,
	})
	if err != nil {
		return nil, err
	}
	data := make([]types.SysCronJobLogItem, 0)
	for _, item := range result.Data {
		data = append(data, types.SysCronJobLogItem{
			Id:             item.Id,
			JobId:          item.JobId,
			JobName:        item.JobName,
			InvokeTarget:   item.InvokeTarget,
			CronExpression: item.CronExpression,
			Status:         item.Status,
			Message:        item.Message,
			ExceptionInfo:  item.ExceptionInfo,
			StartTime:      item.StartTime,
			EndTime:        item.EndTime,
			CreateTime:     item.CreateTime,
		})
	}
	resp = &types.SysCronJobLogListResp{
		RespBase: types.RespBase{
			Code:       result.Base.Code,
			Msg:        result.Base.Msg,
			Total:      result.Base.Total,
			HasNext:    result.Base.HasNext,
			HasPrev:    result.Base.HasPrev,
			NextCursor: result.Base.NextCursor,
			PrevCursor: result.Base.PrevCursor,
		},
		Data: data,
	}
	return
}
