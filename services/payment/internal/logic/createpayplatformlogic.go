package logic

import (
	"context"
	"database/sql"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreatePayPlatformLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreatePayPlatformLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePayPlatformLogic {
	return &CreatePayPlatformLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建平台
func (l *CreatePayPlatformLogic) CreatePayPlatform(in *payment.CreatePayPlatformReq) (*payment.AdminCommonResp, error) {
	var (
		errLogic = "CreatePayPlatform"
	)

	now := utils.NowMillis()
	platform := &models.TPayPlatform{
		PlatformCode: in.PlatformCode,
		PlatformName: in.PlatformName,
		PlatformType: int64(in.PlatformType),
		NotifyUrl:    sql.NullString{String: in.NotifyUrl, Valid: true},
		ReturnUrl:    sql.NullString{String: in.ReturnUrl, Valid: true},
		Icon:         sql.NullString{String: in.Icon, Valid: true},
		Status:       int64(in.Status),
		Remark:       sql.NullString{String: in.Remark, Valid: true},
		CreateTimes:  now,
		UpdateTimes:  now,
	}

	_, err := l.svcCtx.PayPlatformModel.Insert(l.ctx, platform)
	if err != nil {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	l.Logger.Infof("Create pay platform success: %s", in.PlatformCode)

	return &payment.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
