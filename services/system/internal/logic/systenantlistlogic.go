package logic

import (
	"context"

	"wklive/common/pageutil"
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
	items, total, err := l.svcCtx.TenantMode.FindPage(l.ctx, in.Keyword, commonStatusToModel(in.Status), in.TenantName, in.TenantCode, in.ContactName, in.ContactPhone, in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}
	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}

	data := make([]*system.SysTenantItem, 0)
	for _, v := range items {
		data = append(data, &system.SysTenantItem{
			Id:           v.Id,
			TenantCode:   v.TenantCode,
			TenantName:   v.TenantName,
			Status:       commonStatusToProto(v.Status),
			ExpireTime:   v.ExpireTime,
			ContactName:  v.ContactName.String,
			ContactPhone: v.ContactPhone.String,
			Remark:       v.Remark.String,
			CreateTimes:  v.CreateTimes,
			UpdateTimes:  v.UpdateTimes,
		})
	}
	return &system.SysTenantListResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastID),
		Data: data,
	}, nil
}
