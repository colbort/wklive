package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSecurityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSecurityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSecurityLogic {
	return &GetSecurityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 安全设置相关接口
func (l *GetSecurityLogic) GetSecurity(in *user.GetSecurityReq) (*user.GetSecurityResp, error) {
	// todo: add your logic here and delete this line

	return &user.GetSecurityResp{}, nil
}
