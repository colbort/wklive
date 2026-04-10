package logic

import (
	"context"
	"database/sql"
	"time"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTenantPayChannelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTenantPayChannelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTenantPayChannelLogic {
	return &CreateTenantPayChannelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建租户支付通道
func (l *CreateTenantPayChannelLogic) CreateTenantPayChannel(in *payment.CreateTenantPayChannelReq) (*payment.AdminCommonResp, error) {
	var (
		errLogic = "CreateTenantPayChannel"
	)

	feeRate, err := conv.ParseFloatField(in.FeeRate)
	if err != nil {
		return nil, err
	}

	now := time.Now().UnixMilli()
	channel := &models.TTenantPayChannel{
		TenantId:        in.TenantId,
		PlatformId:      in.PlatformId,
		ProductId:       in.ProductId,
		AccountId:       in.AccountId,
		ChannelCode:     in.ChannelCode,
		ChannelName:     in.ChannelName,
		DisplayName:     sql.NullString{String: in.DisplayName, Valid: true},
		Icon:            sql.NullString{String: in.Icon, Valid: true},
		Currency:        in.Currency,
		Sort:            in.Sort,
		Visible:         in.Visible,
		Status:          int64(in.Status),
		SingleMinAmount: in.SingleMinAmount,
		SingleMaxAmount: in.SingleMaxAmount,
		DailyMaxAmount:  in.DailyMaxAmount,
		DailyMaxCount:   in.DailyMaxCount,
		FeeType:         int64(in.FeeType),
		FeeRate:         feeRate,
		FeeFixedAmount:  in.FeeFixedAmount,
		ExtConfig:       sql.NullString{String: in.ExtConfig, Valid: true},
		Remark:          sql.NullString{String: in.Remark, Valid: true},
		CreateTimes:     now,
		UpdateTimes:     now,
	}

	_, err = l.svcCtx.TenantPayChannelModel.Insert(l.ctx, channel)
	if err != nil {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	l.Logger.Infof("Create tenant pay channel success: %s", in.ChannelCode)

	return &payment.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
