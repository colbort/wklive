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

type SysTenantListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysTenantListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysTenantListLogic {
	return &SysTenantListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysTenantListLogic) SysTenantList(req *types.SysTenantListReq) (resp *types.SysTenantListResp, err error) {
	result, err := l.svcCtx.SystemCli.SysTenantList(l.ctx, &system.SysTenantListReq{
		Page: &common.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
		Keyword:      req.Keyword,
		Status:       req.Status,
		TenantName:   req.TenantName,
		TenantCode:   req.TenantCode,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
	})
	if err != nil {
		return nil, err
	}
	data := make([]types.SysTenantItem, 0)
	for _, v := range result.Data {
		data = append(data, types.SysTenantItem{
			Id:           v.Id,
			TenantCode:   v.TenantCode,
			TenantName:   v.TenantName,
			Status:       v.Status,
			ExpireTime:   v.ExpireTime,
			ContactName:  v.ContactName,
			ContactPhone: v.ContactPhone,
			Remark:       v.Remark,
			CreateTimes:   v.CreateTimes,
			UpdateTimes:   v.UpdateTimes,
		})
	}
	return &types.SysTenantListResp{
		RespBase: types.RespBase{
			Code:       result.Base.Code,
			Msg:        result.Base.Msg,
			HasNext:    result.Base.HasNext,
			HasPrev:    result.Base.HasPrev,
			NextCursor: result.Base.NextCursor,
			PrevCursor: result.Base.PrevCursor,
		},
		Data: data,
	}, nil
}
