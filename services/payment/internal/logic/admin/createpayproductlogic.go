package adminlogic

import (
	"context"
	"database/sql"
	"errors"
	"wklive/services/payment/internal/logic/helpers"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreatePayProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreatePayProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePayProductLogic {
	return &CreatePayProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建产品
func (l *CreatePayProductLogic) CreatePayProduct(in *payment.CreatePayProductReq) (*payment.CommonResp, error) {
	var (
		errLogic = "CreatePayProduct"
	)
	if base, err := helpers.SystemAdminWriteScopeResp(l.ctx); err != nil {
		return nil, err
	} else if base != nil {
		return &payment.CommonResp{Base: base}, nil
	}
	if in.PlatformId <= 0 || !requiredStrings(in.ProductCode, in.ProductName, in.Currency) || in.SceneType == 0 {
		return paymentErrorResp(l.ctx, i18n.PaymentRequiredParamsMissing), nil
	}
	if _, ok := payment.SceneType_name[int32(in.SceneType)]; !ok {
		return paymentErrorResp(l.ctx, i18n.ParamError), nil
	}
	if platform, err := l.svcCtx.PayPlatformModel.FindOne(l.ctx, in.PlatformId); errors.Is(err, models.ErrNotFound) || platform == nil {
		return paymentErrorResp(l.ctx, i18n.PlatformNotFound), nil
	} else if err != nil {
		return nil, err
	}

	now := utils.NowMillis()
	product := &models.TPayProduct{
		PlatformId:  in.PlatformId,
		ProductCode: in.ProductCode,
		ProductName: in.ProductName,
		SceneType:   int64(in.SceneType),
		Currency:    in.Currency,
		Enabled:     helpers.EnableToModel(in.Enabled, int64(common.Enable_ENABLE_ENABLED)),
		Remark:      sql.NullString{String: in.Remark, Valid: true},
		CreateTimes: now,
		UpdateTimes: now,
	}

	_, err := l.svcCtx.PayProductModel.Insert(l.ctx, product)
	if err != nil {
		if helpers.IsDuplicateEntry(err) {
			return &payment.CommonResp{
				Base: helper.ErrResp(
					i18n.PayProductCodeAlreadyExists,
					i18n.Translate(i18n.PayProductCodeAlreadyExists, l.ctx),
				),
			}, nil
		}
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	l.Logger.Infof("Create pay product success: %s", in.ProductCode)

	return &payment.CommonResp{
		Base: helper.OkResp(),
	}, nil
}
