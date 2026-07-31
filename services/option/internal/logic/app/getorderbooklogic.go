package applogic

import (
	"context"
	"errors"
	"time"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/internal/observability"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type GetOrderBookLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrderBookLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderBookLogic {
	return &GetOrderBookLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取来自 Option 活动限价委托的聚合盘口
func (l *GetOrderBookLogic) GetOrderBook(in *option.GetOrderBookReq) (*option.GetOrderBookResp, error) {
	if in == nil || in.ContractId <= 0 || in.DepthLimit < 0 || in.DepthLimit > 100 {
		return orderBookParamError(l.ctx), nil
	}
	limit := int64(in.DepthLimit)
	if limit == 0 {
		limit = 20
	}
	tenantID, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	var bids, asks []*models.OptionOrderBookLevel
	var sequence int64
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		contractModel := models.NewTOptionContractModel(conn, l.svcCtx.Config.CacheRedis)
		orderModel := models.NewTOptionOrderModel(conn, l.svcCtx.Config.CacheRedis)
		tradeModel := models.NewTOptionTradeModel(conn, l.svcCtx.Config.CacheRedis)
		if _, findErr := contractModel.FindOneForPublicMarket(ctx, tenantID, in.ContractId); findErr != nil {
			return findErr
		}
		var queryErr error
		bids, queryErr = orderModel.FindOrderBookLevels(
			ctx, tenantID, in.ContractId, int64(common.Side_SIDE_BUY), limit,
		)
		if queryErr != nil {
			return queryErr
		}
		asks, queryErr = orderModel.FindOrderBookLevels(
			ctx, tenantID, in.ContractId, int64(common.Side_SIDE_SELL), limit,
		)
		if queryErr != nil {
			return queryErr
		}
		if simpleBookIsolationViolation(bids) || simpleBookIsolationViolation(asks) {
			observability.RecordComboIsolationViolation(tenantID, "public_book")
			l.Errorf(
				"option combo isolation violation path=public_book tenantId=%d contractId=%d",
				tenantID, in.ContractId,
			)
			return errors.New("option public order book contains combo shadow orders")
		}
		sequence, queryErr = tradeModel.FindLastMatchSequence(ctx, tenantID, in.ContractId)
		return queryErr
	})
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.GetOrderBookResp{
				Base: helper.ErrResp(i18n.ContractNotFound, i18n.Translate(i18n.ContractNotFound, l.ctx)),
			}, nil
		}
		return nil, err
	}
	return &option.GetOrderBookResp{
		Base: helper.OkResp(),
		Data: &option.OptionOrderBook{
			ContractId:        in.ContractId,
			LastMatchSequence: sequence,
			GeneratedAt:       now,
			Source:            "OPTION_ACTIVE_LIMIT_ORDERS",
			Bids:              toOrderBookProto(bids),
			Asks:              toOrderBookProto(asks),
		},
	}, nil
}

func simpleBookIsolationViolation(items []*models.OptionOrderBookLevel) bool {
	for _, item := range items {
		if item != nil && item.ComboOrderCount > 0 {
			return true
		}
	}
	return false
}

func orderBookParamError(ctx context.Context) *option.GetOrderBookResp {
	return &option.GetOrderBookResp{
		Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, ctx)),
	}
}

func toOrderBookProto(items []*models.OptionOrderBookLevel) []*option.OptionOrderBookLevel {
	result := make([]*option.OptionOrderBookLevel, 0, len(items))
	for _, item := range items {
		result = append(result, &option.OptionOrderBookLevel{
			Price:      conv.FloatString(item.Price),
			Qty:        conv.FloatString(item.Qty),
			OrderCount: item.OrderCount,
		})
	}
	return result
}
