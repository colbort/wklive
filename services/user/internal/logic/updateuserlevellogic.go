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

type UpdateUserLevelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserLevelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLevelLogic {
	return &UpdateUserLevelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新用户会员等级
func (l *UpdateUserLevelLogic) UpdateUserLevel(in *user.UpdateUserLevelReq) (*user.AdminCommonResp, error) {
	// 获取用户信息
	tuser, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if tuser == nil {
		return &user.AdminCommonResp{
			Base: &common.RespBase{Code: 404, Msg: "用户不存在"},
		}, nil
	}

	// 更新会员等级
	err = l.svcCtx.UserModel.Update(l.ctx, &models.TUser{
		Id:          in.UserId,
		MemberLevel: in.MemberLevel,
		UpdateTimes: time.Now().UnixMilli(),
	})
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("管理员更新用户 %d 会员等级为 %d", in.UserId, in.MemberLevel)

	return &user.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
