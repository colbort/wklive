package validation

import (
	"errors"
	"wklive/proto/common"

	"wklive/proto/trade"
	"wklive/services/trade/models"
)

func SymbolTradingTimeline(productType common.ProductType, contractType common.ContractType, listingTime, tradingStartTime, tradingEndTime int64) error {
	if listingTime < 0 || tradingStartTime < 0 || tradingEndTime < 0 {
		return errors.New("交易时间不能为负数")
	}
	if listingTime > 0 && tradingStartTime > 0 && tradingStartTime <= listingTime {
		return errors.New("开始交易时间必须晚于上线时间")
	}
	if tradingStartTime > 0 && tradingEndTime > 0 && tradingEndTime <= tradingStartTime {
		return errors.New("停止交易时间必须晚于开始交易时间")
	}
	if listingTime > 0 && tradingEndTime > 0 && tradingEndTime <= listingTime {
		return errors.New("停止交易时间必须晚于上线时间")
	}
	if productType == common.ProductType_PRODUCT_TYPE_DERIVATIVE && contractType == common.ContractType_CONTRACT_TYPE_DELIVERY {
		if listingTime == 0 || tradingStartTime == 0 || tradingEndTime == 0 {
			return errors.New("交割合约必须配置上线、开始交易和停止交易时间")
		}
	}
	return nil
}

func ContractTradingTimeline(symbol *models.TTradeSymbol, deliveryTime, openCutoffTime, matchingStopTime int64) error {
	if symbol == nil || symbol.ProductType != int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE) {
		return errors.New("只有衍生品交易对可以配置合约参数")
	}
	contractType := common.ContractType(symbol.ContractType)
	if contractType == common.ContractType_CONTRACT_TYPE_PERPETUAL {
		if deliveryTime != 0 || openCutoffTime != 0 || matchingStopTime != 0 {
			return errors.New("永续合约不能配置交割、停止开仓或停止撮合时间")
		}
		return nil
	}
	if contractType != common.ContractType_CONTRACT_TYPE_DELIVERY {
		return errors.New("交易对合约类型无效")
	}
	if err := SymbolTradingTimeline(common.ProductType(symbol.ProductType), contractType, symbol.ListingTime, symbol.TradingStartTime, symbol.TradingEndTime); err != nil {
		return err
	}
	if deliveryTime <= 0 || openCutoffTime <= 0 || matchingStopTime <= 0 {
		return errors.New("交割合约必须配置停止开仓、停止撮合和交割时间")
	}
	if openCutoffTime <= symbol.TradingStartTime {
		return errors.New("停止开仓时间必须晚于开始交易时间")
	}
	if matchingStopTime <= openCutoffTime {
		return errors.New("停止撮合时间必须晚于停止开仓时间")
	}
	if matchingStopTime > symbol.TradingEndTime {
		return errors.New("停止撮合时间不能晚于停止交易时间")
	}
	if deliveryTime <= symbol.TradingEndTime {
		return errors.New("交割时间必须晚于停止交易时间")
	}
	return nil
}

func ContractMarginModes(supportCross, supportIsolated int64) error {
	if (supportCross != 0 && supportCross != 1) || (supportIsolated != 0 && supportIsolated != 1) {
		return errors.New("保证金模式支持开关只能为0或1")
	}
	if supportCross == 0 && supportIsolated == 0 {
		return errors.New("全仓和逐仓至少需要支持一种保证金模式")
	}
	return nil
}

func ContractSupportsMarginMode(config *models.TTradeSymbolContract, marginMode trade.MarginMode) bool {
	if config == nil {
		return false
	}
	switch marginMode {
	case trade.MarginMode_MARGIN_MODE_CROSS:
		return config.SupportCross == 1
	case trade.MarginMode_MARGIN_MODE_ISOLATED:
		return config.SupportIsolated == 1
	default:
		return false
	}
}
