package trade

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/trade"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTradeOptionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTradeOptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTradeOptionsLogic {
	return &GetTradeOptionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTradeOptionsLogic) GetTradeOptions() (resp *types.GetTradeOptionsResp, err error) {
	return &types.GetTradeOptionsResp{
		RespBase: types.RespBase{Code: 200, Msg: "success"},
		Data: []types.OptionsGroup{
			logicutil.EnumGroup("marketType", "市场类型", trade.MarketType_MARKET_TYPE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("contractType", "合约类型", trade.ContractType_CONTRACT_TYPE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("symbolStatus", "交易对状态", trade.SymbolStatus_SYMBOL_STATUS_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("tradeSide", "买卖方向", trade.TradeSide_TRADE_SIDE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("orderType", "订单类型", trade.OrderType_ORDER_TYPE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("triggerKind", "触发用途", trade.TriggerKind_TRIGGER_KIND_NONE.Descriptor()),
			logicutil.EnumGroup("timeInForce", "有效方式", trade.TimeInForce_TIME_IN_FORCE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("orderStatus", "订单状态", trade.OrderStatus_ORDER_STATUS_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("eventStatus", "事件状态", trade.EventStatus_EVENT_STATUS_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("marginMode", "保证金模式", trade.MarginMode_MARGIN_MODE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("positionMode", "持仓模式", trade.PositionMode_POSITION_MODE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("enableStatus", "启用状态", trade.EnableStatus_ENABLE_STATUS_DISABLED.Descriptor()),
			logicutil.Group(
				"leverageValue",
				"杠杆倍数",
				logicutil.Option(1, "LEVERAGE_VALUE_1X"),
				logicutil.Option(2, "LEVERAGE_VALUE_2X"),
				logicutil.Option(5, "LEVERAGE_VALUE_5X"),
				logicutil.Option(10, "LEVERAGE_VALUE_10X"),
				logicutil.Option(20, "LEVERAGE_VALUE_20X"),
				logicutil.Option(50, "LEVERAGE_VALUE_50X"),
				logicutil.Option(75, "LEVERAGE_VALUE_75X"),
				logicutil.Option(100, "LEVERAGE_VALUE_100X"),
				logicutil.Option(125, "LEVERAGE_VALUE_125X"),
			),
		},
	}, nil
}
