// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_token

import (
	"context"

	"chat-api/internal/logicutil"
	"chat-api/internal/svc"
	"chat-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOptionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOptionsLogic {
	return &GetOptionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOptionsLogic) GetOptions(req *types.OptionsReq) (resp *types.OptionsResp, err error) {
	return &types.OptionsResp{
		RespBase: types.RespBase{
			Code: 200,
			Msg:  "获取成功",
		},
		Data: types.Options{
			Options: logicutil.CoreOptions(),
		},
	}, nil
}
