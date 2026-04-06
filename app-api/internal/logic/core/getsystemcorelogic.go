// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package core

import (
	"context"
	"encoding/json"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/proto/itick"
	"wklive/proto/system"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSystemCoreLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSystemCoreLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSystemCoreLogic {
	return &GetSystemCoreLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSystemCoreLogic) GetSystemCore() (resp *types.GetSystemCoreResp, err error) {
	key := system.SysConfigType_OBJECT_STORAGE
	cd, err := l.svcCtx.SystemCli.SysConfigDetail(l.ctx, &system.SysConfigDetailReq{
		ConfigKey: &key,
	})
	if err != nil {
		return nil, err
	}
	var config system.SystemCore
	err = json.Unmarshal([]byte(cd.Data.ConfigValue), &config)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.ItickCli.GetKlineIntervals(l.ctx, &itick.AppEmpty{})
	if err != nil {
		return nil, err
	}
	intervals := make([]types.Interval, 0)
	for _, item := range result.Data {
		intervals = append(intervals, types.Interval{
			Name:  item.Name,
			KType: item.KType,
		})
	}
	data := types.SystemCore{
		IsCaptchaEnabled:  config.IsCaptchaEnabled,
		IsRegisterEnabled: config.IsRegisterEnabled,
		IsGuestEnabled:    config.IsGuestEnabled,
		IsCryptoEnabled:   config.IsCryptoEnabled,
		Intervals:         intervals,
	}
	return &types.GetSystemCoreResp{
		RespBase: types.RespBase{
			Code: 200,
			Msg:  "获取系统配置成功",
		},
		Data: data,
	}, nil
}
