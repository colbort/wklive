package system

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	pbsystem "wklive/proto/system"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSystemOptionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSystemOptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSystemOptionsLogic {
	return &GetSystemOptionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSystemOptionsLogic) GetSystemOptions() (resp *types.GetSystemOptionsResp, err error) {
	return &types.GetSystemOptionsResp{
		RespBase: types.RespBase{Code: 200, Msg: "success"},
		Data: []types.OptionsGroup{
			logicutil.EnumGroup("sysConfigType", "系统配置类型", pbsystem.SysConfigType_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("status", "状态", pbsystem.CommonStatus_COMMON_STATUS_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("menuType", "菜单类型", pbsystem.MenuType_MENU_TYPE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("visible", "可见状态", pbsystem.VisibleStatus_VISIBLE_STATUS_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("method", "请求方法", pbsystem.RequestMethod_REQUEST_METHOD_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("jobStatus", "任务状态", pbsystem.JobStatus_JOB_STATUS_DISABLED.Descriptor()),
		},
	}, nil
}
