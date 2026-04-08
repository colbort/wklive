package logic

import (
	"context"

	"wklive/proto/common"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysConfigKeysLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysConfigKeysLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysConfigKeysLogic {
	return &SysConfigKeysLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取系统配置所有的key
func (l *SysConfigKeysLogic) SysConfigKeys(in *system.Empty) (*system.SysConfigKeysResp, error) {
	data := make([]string, 0)
	for _, v := range system.SysConfigType_name {
		if v == system.SysConfigType_UNKNOWN.String() {
			continue
		}
		data = append(data, v)
	}

	return &system.SysConfigKeysResp{
		Base: &common.RespBase{
			Code: 200,
			Msg:  "",
		},
		Data: data,
	}, nil
}
