package logic

import (
	"context"

	"wklive/common/utils"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysConfigListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysConfigListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysConfigListLogic {
	return &SysConfigListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取系统配置列表
func (l *SysConfigListLogic) SysConfigList(in *system.SysConfigListReq) (*system.SysConfigListResp, error) {
	configs, count, err := l.svcCtx.ConfigModel.FindPage(l.ctx, "", in.Page.Page, in.Page.Size)
	if err != nil {
		return nil, err
	}
	var data []*system.SysConfigItem
	for _, config := range configs {
		value, err := utils.StringToStruct(config.ConfigValue.String)
		if err != nil {
			return nil, err
		}
		data = append(data, &system.SysConfigItem{
			Id:          config.Id,
			ConfigKey:   config.ConfigKey.String,
			ConfigValue: value,
			Remark:      config.Remark.String,
			CreatedAt:   config.CreatedAt.Unix(),
		})
	}

	return &system.SysConfigListResp{
		Base: &system.RespBase{
			Code:  200,
			Msg:   "查询成功",
			Total: count,
		},
		Data: data,
	}, nil
}
