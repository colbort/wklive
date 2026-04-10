package logic

import (
	"context"
	"database/sql"
	"time"

	"wklive/common/helper"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTenantPayAccountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTenantPayAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTenantPayAccountLogic {
	return &CreateTenantPayAccountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 租户支付账号
func (l *CreateTenantPayAccountLogic) CreateTenantPayAccount(in *payment.CreateTenantPayAccountReq) (*payment.AdminCommonResp, error) {
	var (
		errLogic = "CreateTenantPayAccount"
	)

	now := time.Now().UnixMilli()
	account := &models.TTenantPayAccount{
		TenantId:            in.TenantId,
		TenantPayPlatformId: in.TenantPayPlatformId,
		PlatformId:          in.PlatformId,
		AccountCode:         in.AccountCode,
		AccountName:         in.AccountName,
		AppId:               sql.NullString{String: in.AppId, Valid: true},
		MerchantId:          sql.NullString{String: in.MerchantId, Valid: true},
		MerchantName:        sql.NullString{String: in.MerchantName, Valid: true},
		PublicKey:           sql.NullString{String: in.PublicKey, Valid: true},
		ExtConfig:           sql.NullString{String: in.ExtConfig, Valid: true},
		Status:              int64(in.Status),
		IsDefault:           in.IsDefault,
		Remark:              sql.NullString{String: in.Remark, Valid: true},
		CreateTimes:         now,
		UpdateTimes:         now,
	}

	_, err := l.svcCtx.TenantPayAccountModel.Insert(l.ctx, account)
	if err != nil {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	l.Logger.Infof("Create tenant pay account success: %s", in.AccountCode)

	return &payment.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
