package logic

import (
	"context"

	"wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysRoleListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysRoleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysRoleListLogic {
	return &SysRoleListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 角色
func (l *SysRoleListLogic) SysRoleList(in *system.SysRoleListReq) (*system.SysRoleListResp, error) {
	tenantId, _ := utils.GetTenantIdFromMd(l.ctx)
	// 2) 查分页
	items, total, err := l.svcCtx.RoleModel.FindPage(l.ctx, in.Keyword, tenantId, commonStatusToModel(in.Status), in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}

	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}

	// 3) 组装返回
	data := make([]*system.SysRoleItem, 0, len(items))
	for _, r := range items {
		if r.Code == "super_admin" || r.Code == "tenant_super_admin" {
			continue
		}
		data = append(data, &system.SysRoleItem{
			Id:          r.Id,
			Name:        r.Name,
			Code:        r.Code,
			Status:      commonStatusToProto(r.Status),
			TenantId:    r.TenantId,
			Remark:      r.Remark,
			CreateTimes: r.CreateTimes,
		})
	}

	return &system.SysRoleListResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastID),
		Data: data,
	}, nil
}
