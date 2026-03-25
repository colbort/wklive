package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetIdentityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetIdentityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetIdentityLogic {
	return &GetIdentityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 实名认证相关接口
func (l *GetIdentityLogic) GetIdentity(in *user.GetIdentityReq) (*user.GetIdentityResp, error) {
	// todo: add your logic here and delete this line

	return &user.GetIdentityResp{}, nil
}
