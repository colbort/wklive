// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateRiskLevelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateRiskLevelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRiskLevelLogic {
	return &UpdateRiskLevelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateRiskLevelLogic) UpdateRiskLevel(req *types.UpdateRiskLevelReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.UserCli.UpdateRiskLevel(l.ctx, &user.UpdateRiskLevelReq{
		TenantId:  req.TenantId,
		UserId:    req.UserId,
		RiskLevel: user.RiskLevel(req.RiskLevel),
	})
	if err != nil {
		return nil, err
	}

	return &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}, nil
}
