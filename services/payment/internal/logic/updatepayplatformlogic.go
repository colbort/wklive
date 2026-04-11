package logic

import (
	"context"
	"database/sql"
	"github.com/zeromicro/go-zero/core/logx"
	"time"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
)

type UpdatePayPlatformLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdatePayPlatformLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePayPlatformLogic {
	return &UpdatePayPlatformLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新平台
func (l *UpdatePayPlatformLogic) UpdatePayPlatform(in *payment.UpdatePayPlatformReq) (*payment.AdminCommonResp, error) {
	var (
		errLogic = "UpdatePayPlatform"
	)

	// 査询平台是否存在
	platform, err := l.svcCtx.PayPlatformModel.FindOne(l.ctx, in.Id)
	if err != nil {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if platform == nil {
		return &payment.AdminCommonResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.PlatformNotFound, l.ctx)),
		}, nil
	}

	now := time.Now().UnixMilli()
	if in.PlatformName != "" {
		platform.PlatformName = in.PlatformName
	}
	if in.PlatformType != 0 {
		platform.PlatformType = int64(in.PlatformType)
	}
	if in.NotifyUrl != "" {
		platform.NotifyUrl = sql.NullString{String: in.NotifyUrl, Valid: true}
	}
	if in.ReturnUrl != "" {
		platform.ReturnUrl = sql.NullString{String: in.ReturnUrl, Valid: true}
	}
	if in.Icon != "" {
		platform.Icon = sql.NullString{String: in.Icon, Valid: true}
	}
	if in.Status != 0 {
		platform.Status = int64(in.Status)
	}
	if in.Remark != "" {
		platform.Remark = sql.NullString{String: in.Remark, Valid: true}
	}
	platform.UpdateTimes = now

	err = l.svcCtx.PayPlatformModel.Update(l.ctx, platform)
	if err != nil {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	l.Logger.Infof("Update pay platform success: %d", in.Id)

	return &payment.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
