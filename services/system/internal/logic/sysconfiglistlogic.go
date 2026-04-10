package logic

import (
	"context"

	"wklive/common/pageutil"
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
	items, total, err := l.svcCtx.ConfigModel.FindPage(l.ctx, "", in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}

	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}

	var data []*system.SysConfigItem
	for _, config := range items {
		data = append(data, &system.SysConfigItem{
			Id:          config.Id,
			ConfigKey:   config.ConfigKey.String,
			ConfigValue: config.ConfigValue.String,
			Remark:      config.Remark.String,
			CreateTimes: config.CreateTimes,
		})
	}

	return &system.SysConfigListResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastID),
		Data: data,
	}, nil
}
