package logic

import (
	"context"

	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysConfigDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysConfigDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysConfigDeleteLogic {
	return &SysConfigDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除系统配置
func (l *SysConfigDeleteLogic) SysConfigDelete(in *system.SysConfigDeleteReq) (*system.RespBase, error) {
	err := l.svcCtx.ConfigModel.Delete(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &system.RespBase{
		Code: 200,
		Msg:  "删除成功",
	}, nil
}
