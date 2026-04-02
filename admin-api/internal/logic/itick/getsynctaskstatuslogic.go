// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/itick"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSyncTaskStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSyncTaskStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSyncTaskStatusLogic {
	return &GetSyncTaskStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSyncTaskStatusLogic) GetSyncTaskStatus(req *types.GetSyncTaskStatusReq) (resp *types.GetSyncTaskStatusResp, err error) {
	result, err := l.svcCtx.ItickCli.GetSyncTaskStatus(l.ctx, &itick.GetSyncTaskStatusReq{
		TaskNo: req.TaskNo,
	})
	if err != nil {
		return nil, err
	}
	return &types.GetSyncTaskStatusResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		TaskNo:  result.TaskNo,
		Status:  result.Status,
		Message: result.Message,
	}, nil
}
