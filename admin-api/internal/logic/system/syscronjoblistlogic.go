// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/common"
	"wklive/proto/system"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysCronJobListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysCronJobListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysCronJobListLogic {
	return &SysCronJobListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysCronJobListLogic) SysCronJobList(req *types.SysCronJobListReq) (resp *types.SysCronJobListResp, err error) {
	result, err := l.svcCtx.SystemCli.SysCronJobList(l.ctx, &system.SysCronJobListReq{
		Page: &common.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
		Keyword:  req.Keyword,
		JobName:  req.JobName,
		JobGroup: req.JobGroup,
		Status:   req.Status,
	})
	if err != nil {
		return nil, err
	}
	data := make([]types.SysCronJobItem, 0)
	for _, item := range result.Data {
		data = append(data, types.SysCronJobItem{
			Id:             item.Id,
			JobName:        item.JobName,
			JobGroup:       item.JobGroup,
			InvokeTarget:   item.InvokeTarget,
			CronExpression: item.CronExpression,
			Status:         item.Status,
			Remark:         item.Remark,
			CreateBy:       item.CreateBy,
			CreateTimes:     item.CreateTimes,
			UpdateBy:       item.UpdateBy,
			UpdateTimes:     item.UpdateTimes,
		})
	}
	resp = &types.SysCronJobListResp{
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
