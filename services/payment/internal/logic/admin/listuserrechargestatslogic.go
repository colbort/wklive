package adminlogic

import (
	"context"
	"errors"
	"strings"

	"wklive/common/i18n"
	"wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserRechargeStatsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUserRechargeStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserRechargeStatsLogic {
	return &ListUserRechargeStatsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 用户充值统计列表
func (l *ListUserRechargeStatsLogic) ListUserRechargeStats(in *payment.ListUserRechargeStatsReq) (*payment.ListUserRechargeStatsResp, error) {
	if in.TenantId <= 0 {
		if tenantId, err := utils.GetTenantIdFromMd(l.ctx); err == nil {
			in.TenantId = tenantId
		}
	}
	minAmountText := ""
	minAmount, err := parseNonNegativeAmount(in.SuccessTotalAmountMin)
	if strings.TrimSpace(in.SuccessTotalAmountMin) != "" {
		if err != nil {
			return &payment.ListUserRechargeStatsResp{
				Base: paymentErrorResp(l.ctx, i18n.InvalidPaymentDecimal).Base,
			}, nil
		}
		minAmountText = minAmount.String()
	}
	maxAmountText := ""
	maxAmount, err := parseNonNegativeAmount(in.SuccessTotalAmountMax)
	if strings.TrimSpace(in.SuccessTotalAmountMax) != "" {
		if err != nil {
			return &payment.ListUserRechargeStatsResp{
				Base: paymentErrorResp(l.ctx, i18n.InvalidPaymentDecimal).Base,
			}, nil
		}
		if minAmount.GreaterThan(maxAmount) {
			return &payment.ListUserRechargeStatsResp{
				Base: paymentErrorResp(l.ctx, i18n.InvalidPaymentAmountRange).Base,
			}, nil
		}
		maxAmountText = maxAmount.String()
	}
	stats, total, err := l.svcCtx.UserRechargeStatModel.FindPage(
		l.ctx,
		models.UserRechargeStatPageFilter{
			TenantId:              in.TenantId,
			UserId:                in.UserId,
			SuccessTotalAmountMin: minAmountText,
			SuccessTotalAmountMax: maxAmountText,
		},
		in.Page.Cursor,
		in.Page.Limit,
	)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(stats) > 0 {
		lastID = stats[len(stats)-1].Id
	}

	data := make([]*payment.UserRechargeStat, 0, len(stats))
	for _, s := range stats {
		data = append(data, toUserRechargeStatProto(s))
	}

	return &payment.ListUserRechargeStatsResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(stats), total, lastID),
		Data: data,
	}, nil
}
