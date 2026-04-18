package itick

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/itick"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetItickOptionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetItickOptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetItickOptionsLogic {
	return &GetItickOptionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetItickOptionsLogic) GetItickOptions() (resp *types.GetItickOptionsResp, err error) {
	return &types.GetItickOptionsResp{
		RespBase: types.RespBase{Code: 200, Msg: "success"},
		Data: []types.OptionsGroup{
			logicutil.EnumGroup("categoryType", "产品类型", itick.CategoryType_CATEGORY_TYPE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("klineType", "K线周期", itick.KlineType_KLINE_TYPE_UNKNOWN.Descriptor()),
			logicutil.Group("status", "状态",
				logicutil.Option(0, "UNKNOWN"),
				logicutil.Option(1, "COMMON_STATUS_ENABLED"),
				logicutil.Option(2, "COMMON_STATUS_DISABLED"),
			),
			logicutil.Group("visible", "显示状态",
				logicutil.Option(0, "UNKNOWN"),
				logicutil.Option(1, "YES_NO_YES"),
				logicutil.Option(2, "YES_NO_NO"),
			),
		},
	}, nil
}
