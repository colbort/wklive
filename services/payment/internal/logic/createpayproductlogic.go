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
func (l *CreatePayProductLogic) CreatePayProduct(in *payment.CreatePayProductReq) (*payment.AdminCommonResp, error) {
	var (
		errLogic = "CreatePayProduct"
	)

	now := utils.NowMillis()
	product := &models.TPayProduct{
		PlatformId:  in.PlatformId,
		ProductCode: in.ProductCode,
		ProductName: in.ProductName,
		SceneType:   int64(in.SceneType),
		Currency:    in.Currency,
		Status:      int64(in.Status),
		Remark:      sql.NullString{String: in.Remark, Valid: true},
		CreateTimes: now,
		UpdateTimes: now,
	}

	_, err := l.svcCtx.PayProductModel.Insert(l.ctx, product)
	if err != nil {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	l.Logger.Infof("Create pay product success: %s", in.ProductCode)

	return &payment.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
