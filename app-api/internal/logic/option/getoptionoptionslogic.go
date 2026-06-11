package option

import (
	"context"

	"wklive/app-api/internal/logicutil"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/proto/common"
	"wklive/proto/option"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOptionOptionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOptionOptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOptionOptionsLogic {
	return &GetOptionOptionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOptionOptionsLogic) GetOptionOptions() (resp *types.GetOptionOptionsResp, err error) {
	return &types.GetOptionOptionsResp{
		RespBase: types.RespBase{Code: 200, Msg: "success"},
		Data: []types.OptionsGroup{
			logicutil.EnumGroup("commonStatus", "通用状态", common.Enable_ENABLE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("optionType", "期权类型", option.OptionType_OPTION_TYPE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("exerciseStyle", "行权方式", option.ExerciseStyle_EXERCISE_STYLE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("contractStatus", "合约状态", option.ContractStatus_CONTRACT_STATUS_UNKNOWN.Descriptor()),
		},
	}, nil
}
