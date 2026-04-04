package logic

import (
	"context"

	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysConfigByKeysLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysConfigByKeysLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysConfigByKeysLogic {
	return &SysConfigByKeysLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取系统配置根据keys
func (l *SysConfigByKeysLogic) SysConfigByKeys(in *system.SysConfigByKeysReq) (*system.SysConfigByKeysResp, error) {
	configs, err := l.svcCtx.ConfigModel.FindByKeys(l.ctx, in.ConfigKeys)
	if err != nil {
		return nil, err
	}

	var data []*system.SysConfigItem
	for _, config := range configs {
		data = append(data, &system.SysConfigItem{
			Id:          config.Id,
			ConfigKey:   config.ConfigKey.String,
			ConfigValue: config.ConfigValue.String,
			Remark:      config.Remark.String,
			CreateTimes: config.CreateTimes,
		})
	}

	return &system.SysConfigByKeysResp{
		Base: &system.RespBase{
			Code: 200,
			Msg:  "查询成功",
		},
		Data: data,
	}, nil
}
