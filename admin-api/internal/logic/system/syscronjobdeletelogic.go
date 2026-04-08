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

type SysCronJobDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysCronJobDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysCronJobDeleteLogic {
	return &SysCronJobDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysCronJobDeleteLogic) SysCronJobDelete(req *types.SysCronJobDeleteReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.SystemCli.SysCronJobDelete(l.ctx, &system.SysCronJobDeleteReq{
		Id: req.Id,
	})
	if err != nil {
		return nil, err
	}
	resp = &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}
	return
}
