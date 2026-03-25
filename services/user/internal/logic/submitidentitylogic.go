package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitIdentityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubmitIdentityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitIdentityLogic {
	return &SubmitIdentityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 提交实名认证信息
func (l *SubmitIdentityLogic) SubmitIdentity(in *user.SubmitIdentityReq) (*user.SubmitIdentityResp, error) {
	// todo: add your logic here and delete this line

	return &user.SubmitIdentityResp{}, nil
}
