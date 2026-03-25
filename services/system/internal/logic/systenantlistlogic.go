package logic

import (
	"context"

	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysTenantListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysTenantListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysTenantListLogic {
	return &SysTenantListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取租户列表
func (l *SysTenantListLogic) SysTenantList(in *system.SysTenantListReq) (*system.SysTenantListResp, error) {
	items, total, err := l.svcCtx.TenantMode.FindPage(l.ctx, in.Keyword, in.Status, in.TenantName, in.TenantCode, in.ContactName, in.ContactPhone, in.Page.Cursor, in.Page.Limit)
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

	data := make([]*system.SysTenantItem, 0)
	for _, v := range items {
		data = append(data, &system.SysTenantItem{
			Id:           v.Id,
			TenantCode:   v.TenantCode,
			TenantName:   v.TenantName,
			Status:       v.Status,
			ExpireTime:   v.ExpireTime.Time.UnixMilli(),
			ContactName:  v.ContactName.String,
			ContactPhone: v.ContactPhone.String,
			Remark:       v.Remark.String,
			CreateTime:   v.CreateTime.UnixMilli(),
			UpdateTime:   v.UpdateTime.UnixMilli(),
		})
	}
	return &system.SysTenantListResp{
		Base: &system.RespBase{
			Code:       200,
			Msg:        "查询成功",
			Total:      total,
			HasNext:    hasNext,
			HasPrev:    hasPrev,
			NextCursor: nextCursor,
			PrevCursor: prevCursor,
		},
		Data: data,
	}, nil
}
