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

type SetDefaultUserBankLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetDefaultUserBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetDefaultUserBankLogic {
	return &SetDefaultUserBankLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置默认用户银行卡
func (l *SetDefaultUserBankLogic) SetDefaultUserBank(in *user.SetDefaultUserBankReq) (*user.AdminCommonResp, error) {
	// 获取要设置为默认的银行卡
	userBank, err := l.svcCtx.UserBankModel.FindOne(l.ctx, in.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if userBank == nil {
		return &user.AdminCommonResp{
			Base: &common.RespBase{Code: 404, Msg: "银行卡不存在"},
		}, nil
	}

	if userBank.UserId != in.UserId {
		return &user.AdminCommonResp{
			Base: &common.RespBase{Code: 403, Msg: "无权限修改"},
		}, nil
	}

	// 如果已经是默认卡，直接返回
	if userBank.IsDefault == 1 {
		return &user.AdminCommonResp{
			Base: helper.OkResp(),
		}, nil
	}

	// 先将该用户的所有默认卡改为非默认
	// TODO: 需要在 UserBankModel 中声明方法，这里暂时跳过

	// 再将指定的卡设置为默认
	err = l.svcCtx.UserBankModel.Update(l.ctx, &models.TUserBank{
		Id:          in.Id,
		IsDefault:   1,
		UpdateTimes: time.Now().UnixMilli(),
	})
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("用户 %d 设置默认银行卡 %d", in.UserId, in.Id)

	return &user.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
