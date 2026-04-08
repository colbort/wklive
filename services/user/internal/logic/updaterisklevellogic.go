package logic

import (
	"context"

	"wklive/proto/common"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateRiskLevelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateRiskLevelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRiskLevelLogic {
	return &UpdateRiskLevelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新用户风险等级
func (l *UpdateRiskLevelLogic) UpdateRiskLevel(in *user.UpdateRiskLevelReq) (*user.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &user.AdminCommonResp{
		Base: &common.RespBase{},
	}, nil
}
