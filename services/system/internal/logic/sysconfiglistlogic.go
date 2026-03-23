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
	items, count, err := l.svcCtx.ConfigModel.FindPage(l.ctx, "", in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}

	prevCursor := in.Page.Cursor
	if prevCursor < 0 {
		prevCursor = 0
	}
	nextCursor := int64(0)
	if int64(len(items)) == in.Page.Limit {
		lastItem := items[len(items)-1]
		nextCursor = lastItem.Id
	}
	hasPrev := prevCursor > 0
	hasNext := int64(len(items)) == in.Page.Limit

	var data []*system.SysConfigItem
	for _, config := range items {
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
			Code:       200,
			Msg:        "查询成功",
			Total:      count,
			HasNext:    hasNext,
			HasPrev:    hasPrev,
			NextCursor: nextCursor,
			PrevCursor: prevCursor,
		},
		Data: data,
	}, nil
}
