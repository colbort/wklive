package logic

import (
	"context"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserStatusLogic {
	return &UpdateUserStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 管理员更新用户状态
func (l *UpdateUserStatusLogic) UpdateUserStatus(in *user.UpdateUserStatusReq) (*user.AdminCommonResp, error) {
	// 获取用户信息
	tuser, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if tuser == nil {
		return &user.AdminCommonResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.UserNotFound, l.ctx)),
		}, nil
	}

	// 更新用户状态
	if in.Status != 0 {
		tuser.Status = int64(in.Status)
	}
	if in.Remark != "" {
		tuser.Remark.String = in.Remark
		tuser.Remark.Valid = true
	}
	tuser.UpdateTimes = utils.NowMillis()

	err = l.svcCtx.UserModel.Update(l.ctx, tuser)
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("管理员更新用户 %d 状态为 %d，备注：%s", in.UserId, in.Status, in.Remark)

	return &user.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
