package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyCryptoRechargeTxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMyCryptoRechargeTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyCryptoRechargeTxLogic {
	return &GetMyCryptoRechargeTxLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 我的链上充值交易详情
func (l *GetMyCryptoRechargeTxLogic) GetMyCryptoRechargeTx(in *payment.GetMyCryptoRechargeTxReq) (*payment.GetMyCryptoRechargeTxResp, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	item, err := l.svcCtx.CryptoRechargeTxModel.FindOneByIdOrHash(l.ctx, tenantId, in.Id, 0, in.TxHash)
	if err != nil {
		if isNotFound(err) {
			return &payment.GetMyCryptoRechargeTxResp{Base: helper.GetErrResp(i18n.CryptoRechargeTxNotFound, i18n.Translate(i18n.CryptoRechargeTxNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if item.UserId != userId {
		return &payment.GetMyCryptoRechargeTxResp{Base: helper.GetErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx))}, nil
	}
	return &payment.GetMyCryptoRechargeTxResp{Base: helper.OkResp(), Data: toCryptoRechargeTxProto(item)}, nil
}
