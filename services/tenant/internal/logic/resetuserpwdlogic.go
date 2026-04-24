package logic

import (
	"context"

	"wklive/proto/tenant"
	"wklive/services/tenant/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResetUserPwdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResetUserPwdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetUserPwdLogic {
	return &ResetUserPwdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 重置租户密码
func (l *ResetUserPwdLogic) ResetUserPwd(in *tenant.ResetUserPwdReq) (*tenant.RespBase, error) {
	// todo: add your logic here and delete this line

	return &tenant.RespBase{}, nil
}
