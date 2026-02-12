package logic

import (
	"context"

	"wklive/rpc/system"
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
	// 1) 分页兜底
	page := in.Page.Page
	if page <= 0 {
		page = 1
	}
	pageSize := in.Page.Size
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	// 2) 查分页
	rows, total, err := l.svcCtx.RoleModel.FindPage(l.ctx, in.Keyword, in.Status, page, pageSize)
	if err != nil {
		return nil, err
	}

	// 3) 组装返回
	data := make([]*system.SysRoleItem, 0, len(rows))
	for _, r := range rows {
		data = append(data, &system.SysRoleItem{
			Id:        r.Id,
			Name:      r.Name,
			Code:      r.Code,
			Status:    int32(r.Status),
			Remark:    r.Remark,
			CreatedAt: r.CreatedAt.UnixMilli(),
		})
	}

	return &system.SysRoleListResp{
		Base: &system.RespBase{
			Code:  200,
			Msg:   "success",
			Total: total,
		},
		Data: data,
	}, nil
}
