package adminlogic

import (
	"context"
	"database/sql"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
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
func (l *CreatePayPlatformLogic) CreatePayPlatform(in *payment.CreatePayPlatformReq) (*payment.CommonResp, error) {
	var (
		errLogic = "CreatePayPlatform"
	)
	if base, err := systemAdminWriteScopeResp(l.ctx); err != nil {
		return nil, err
	} else if base != nil {
		return &payment.CommonResp{Base: base}, nil
	}
	if !requiredStrings(in.PlatformCode, in.PlatformName) || in.PlatformType == 0 {
		return paymentErrorResp(l.ctx, i18n.PaymentRequiredParamsMissing), nil
	}
	if _, ok := payment.PlatformType_name[int32(in.PlatformType)]; !ok {
		return paymentErrorResp(l.ctx, i18n.ParamError), nil
	}

	now := utils.NowMillis()
	platform := &models.TPayPlatform{
		PlatformCode: in.PlatformCode,
		PlatformName: in.PlatformName,
		PlatformType: int64(in.PlatformType),
		NotifyUrl:    sql.NullString{String: in.NotifyUrl, Valid: true},
		ReturnUrl:    sql.NullString{String: in.ReturnUrl, Valid: true},
		Icon:         sql.NullString{String: in.Icon, Valid: true},
		Enabled:      enableToModel(in.Enabled, int64(common.Enable_ENABLE_ENABLED)),
		Remark:       sql.NullString{String: in.Remark, Valid: true},
		CreateTimes:  now,
		UpdateTimes:  now,
	}

	_, err := l.svcCtx.PayPlatformModel.Insert(l.ctx, platform)
	if err != nil {
		if isDuplicateEntry(err) {
			return paymentErrorResp(l.ctx, i18n.PayPlatformCodeAlreadyExists), nil
		}
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	l.Logger.Infof("Create pay platform success: %s", in.PlatformCode)

	return &payment.CommonResp{
		Base: helper.OkResp(),
	}, nil
}
