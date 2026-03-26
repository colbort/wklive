// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPayPlatformsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListPayPlatformsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPayPlatformsLogic {
	return &ListPayPlatformsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListPayPlatformsLogic) ListPayPlatforms(req *types.ListPayPlatformsReq) (resp *types.ListPayPlatformsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
