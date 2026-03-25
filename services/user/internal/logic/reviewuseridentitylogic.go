package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReviewUserIdentityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReviewUserIdentityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewUserIdentityLogic {
	return &ReviewUserIdentityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 审核实名认证信息
func (l *ReviewUserIdentityLogic) ReviewUserIdentity(in *user.ReviewUserIdentityReq) (*user.ReviewUserIdentityResp, error) {
	// todo: add your logic here and delete this line

	return &user.ReviewUserIdentityResp{}, nil
}
