// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/common/utils"
	"wklive/proto/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryMyPayOrderStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryMyPayOrderStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMyPayOrderStatusLogic {
	return &QueryMyPayOrderStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryMyPayOrderStatusLogic) QueryMyPayOrderStatus(req *types.QueryMyPayOrderStatusReq) (resp *types.QueryMyPayOrderStatusResp, err error) {
	userId, err := utils.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.PaymentCli.QueryMyPayOrderStatus(l.ctx, &payment.QueryMyPayOrderStatusReq{
		UserId:   userId,
		TenantId: tenantId,
		OrderNo:  req.OrderNo,
	})
	if err != nil {
		return nil, err
	}
	return &types.QueryMyPayOrderStatusResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: types.PayOrder{
			Id:           result.Order.Id,
			TenantId:     result.Order.TenantId,
			UserId:       result.Order.UserId,
			OrderNo:      result.Order.OrderNo,
			BizOrderNo:   result.Order.BizOrderNo,
			PlatformId:   result.Order.PlatformId,
			ProductId:    result.Order.ProductId,
			AccountId:    result.Order.AccountId,
			ChannelId:    result.Order.ChannelId,
			Currency:     result.Order.Currency,
			OrderAmount:  result.Order.OrderAmount,
			PayAmount:    result.Order.PayAmount,
			FeeAmount:    result.Order.FeeAmount,
			Subject:      result.Order.Subject,
			Body:         result.Order.Body,
			ClientType:   int64(result.Order.ClientType.Number()),
			ClientIp:     result.Order.ClientIp,
			Status:       int64(result.Order.Status.Number()),
			ThirdTradeNo: result.Order.ThirdTradeNo,
			ThirdOrderNo: result.Order.ThirdOrderNo,
			PayUrl:       result.Order.PayUrl,
			QrContent:    result.Order.QrContent,
			RequestData:  result.Order.RequestData,
			ResponseData: result.Order.ResponseData,
			NotifyData:   result.Order.NotifyData,
			ExpireTime:   result.Order.ExpireTime,
			PaidTime:     result.Order.PaidTime,
			NotifyTime:   result.Order.NotifyTime,
			CloseTime:    result.Order.CloseTime,
			Remark:       result.Order.Remark,
			CreateTime:   result.Order.CreateTime,
			UpdateTime:   result.Order.UpdateTime,
		},
	}, nil
}
