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

type UpdateUserBankStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserBankStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserBankStatusLogic {
	return &UpdateUserBankStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新用户银行卡状态
func (l *UpdateUserBankStatusLogic) UpdateUserBankStatus(in *user.UpdateUserBankStatusReq) (*user.AdminCommonResp, error) {
	// 获取银行卡信息
	userBank, err := l.svcCtx.UserBankModel.FindOne(l.ctx, in.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if userBank == nil {
		return &user.AdminCommonResp{
			Base: &common.RespBase{Code: 404, Msg: "银行卡不存在"},
		}, nil
	}

	// 更新银行卡状态
	err = l.svcCtx.UserBankModel.Update(l.ctx, &models.TUserBank{
		Id:          in.Id,
		Status:      int64(in.Status),
		UpdateTimes: time.Now().UnixMilli(),
	})
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("管理员更新银行卡 %d 状态为 %d", in.Id, in.Status)

	return &user.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
