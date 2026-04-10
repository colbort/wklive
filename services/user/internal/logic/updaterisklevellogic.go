package logic

import (
	"context"
	"errors"
	"time"

	"wklive/common/helper"
	"wklive/proto/common"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

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
	// 获取用户安全信息
	userSecurity, err := l.svcCtx.UserSecurityModel.FindOne(l.ctx, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if userSecurity == nil {
		return &user.AdminCommonResp{
			Base: &common.RespBase{Code: 404, Msg: "用户安全信息不存在"},
		}, nil
	}

	// 更新风险等级
	err = l.svcCtx.UserSecurityModel.Update(l.ctx, &models.TUserSecurity{
		Id:          userSecurity.Id,
		RiskLevel:   int64(in.RiskLevel),
		UpdateTimes: time.Now().UnixMilli(),
	})
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("管理员更新用户 %d 风险等级为 %d", in.UserId, in.RiskLevel)

	return &user.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
