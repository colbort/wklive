package applogic

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"wklive/common/conv"
	"wklive/proto/common"
	"wklive/proto/option"
	logichelpers "wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

const (
	minComboLegs  = 2
	maxComboLegs  = 4
	maxComboRatio = 8
)

type validatedComboLeg struct {
	legNo        int64
	contract     *models.TOptionContract
	side         common.Side
	ratio        int64
	price        decimal.Decimal
	qty          decimal.Decimal
	marginCoin   string
	marginAmount decimal.Decimal
}

type validatedComboOrder struct {
	qty                decimal.Decimal
	netPrice           decimal.Decimal
	strategyKey        string
	inverseStrategyKey string
	payloadHash        string
	legs               []*validatedComboLeg
}

type normalizedComboLegInput struct {
	contractID int64
	side       common.Side
	ratio      int64
	price      decimal.Decimal
}

// comboRequestPayloadHash is deliberately independent of current contract,
// market and calendar state so an exact idempotent retry can return the
// original order even after the market has closed.
func comboRequestPayloadHash(in *option.PlaceComboOrderReq) (string, error) {
	if in == nil || in.AccountId <= 0 {
		return "", errors.New("account_id is required")
	}
	clientComboID := strings.TrimSpace(in.ClientComboId)
	if clientComboID == "" || len(clientComboID) > 64 || clientComboID != in.ClientComboId {
		return "", errors.New("client_combo_id must be 1 to 64 non-whitespace bytes")
	}
	if in.OrderType != option.ComboOrderType_COMBO_ORDER_TYPE_LIMIT &&
		in.OrderType != option.ComboOrderType_COMBO_ORDER_TYPE_FOK {
		return "", errors.New("only LIMIT and FOK combo orders are supported")
	}
	if len(in.Legs) < minComboLegs || len(in.Legs) > maxComboLegs {
		return "", fmt.Errorf("combo order must contain %d to %d legs", minComboLegs, maxComboLegs)
	}
	qty, err := conv.ParseDecimalField(in.Qty)
	if err != nil || !validComboDecimal(qty, false) || !qty.IsPositive() {
		return "", errors.New("invalid combo quantity")
	}
	netPrice, err := conv.ParseDecimalField(in.NetPrice)
	if err != nil || !validComboDecimal(netPrice, true) {
		return "", errors.New("invalid combo net price")
	}
	seenContracts := make(map[int64]struct{}, len(in.Legs))
	ratios := make([]int64, 0, len(in.Legs))
	normalizedLegs := make([]normalizedComboLegInput, 0, len(in.Legs))
	hasBuy, hasSell := false, false
	calculatedNet := decimal.Zero
	for _, leg := range in.Legs {
		if leg == nil || leg.ContractId <= 0 {
			return "", errors.New("each combo leg requires a contract")
		}
		if _, exists := seenContracts[leg.ContractId]; exists {
			return "", errors.New("combo legs must use distinct contracts")
		}
		seenContracts[leg.ContractId] = struct{}{}
		if leg.Side != common.Side_SIDE_BUY && leg.Side != common.Side_SIDE_SELL {
			return "", errors.New("each combo leg requires BUY or SELL")
		}
		if leg.PositionEffect != option.PositionEffect_POSITION_EFFECT_OPEN {
			return "", errors.New("combo V1 supports OPEN legs only")
		}
		if leg.Ratio < 1 || leg.Ratio > maxComboRatio {
			return "", fmt.Errorf("each leg ratio must be between 1 and %d", maxComboRatio)
		}
		price, parseErr := conv.ParseDecimalField(leg.Price)
		if parseErr != nil || !validComboDecimal(price, false) || !price.IsPositive() {
			return "", errors.New("invalid combo leg price")
		}
		sign := int64(1)
		if leg.Side == common.Side_SIDE_SELL {
			sign = -1
			hasSell = true
		} else {
			hasBuy = true
		}
		calculatedNet = calculatedNet.Add(
			price.Mul(decimal.NewFromInt(sign * leg.Ratio)),
		)
		ratios = append(ratios, leg.Ratio)
		normalizedLegs = append(normalizedLegs, normalizedComboLegInput{
			contractID: leg.ContractId, side: leg.Side, ratio: leg.Ratio, price: price,
		})
	}
	if !hasBuy || !hasSell {
		return "", errors.New("combo order must contain at least one BUY and one SELL leg")
	}
	if gcdAll(ratios) != 1 {
		return "", errors.New("combo leg ratios must be reduced to lowest terms")
	}
	if !calculatedNet.Equal(netPrice) {
		return "", fmt.Errorf("net_price must equal signed leg-price sum %s", calculatedNet.String())
	}
	sort.Slice(normalizedLegs, func(i, j int) bool {
		return normalizedLegs[i].contractID < normalizedLegs[j].contractID
	})
	strategyParts := make([]string, 0, len(normalizedLegs))
	for _, leg := range normalizedLegs {
		sign := int64(1)
		if leg.side == common.Side_SIDE_SELL {
			sign = -1
		}
		strategyParts = append(
			strategyParts,
			fmt.Sprintf("%d:%+d", leg.contractID, sign*leg.ratio),
		)
	}
	payload := fmt.Sprintf(
		"v1|account=%d|client=%s|type=%d|qty=%s|net=%s|strategy=%s",
		in.AccountId, clientComboID, in.OrderType, qty.String(), netPrice.String(),
		strings.Join(strategyParts, "|"),
	)
	for _, leg := range normalizedLegs {
		payload += fmt.Sprintf("|leg=%d,%d,%d,%s",
			leg.contractID, leg.side, leg.ratio, leg.price.String(),
		)
	}
	payloadSum := sha256.Sum256([]byte(payload))
	return fmt.Sprintf("%x", payloadSum[:]), nil
}

func validateComboOrder(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	tenantID int64,
	in *option.PlaceComboOrderReq,
) (*validatedComboOrder, error) {
	if in == nil || in.AccountId <= 0 {
		return nil, errors.New("account_id is required")
	}
	clientComboID := strings.TrimSpace(in.ClientComboId)
	if clientComboID == "" || len(clientComboID) > 64 || clientComboID != in.ClientComboId {
		return nil, errors.New("client_combo_id must be 1 to 64 non-whitespace bytes")
	}
	if in.OrderType != option.ComboOrderType_COMBO_ORDER_TYPE_LIMIT &&
		in.OrderType != option.ComboOrderType_COMBO_ORDER_TYPE_FOK {
		return nil, errors.New("only LIMIT and FOK combo orders are supported")
	}
	if len(in.Legs) < minComboLegs || len(in.Legs) > maxComboLegs {
		return nil, fmt.Errorf("combo order must contain %d to %d legs", minComboLegs, maxComboLegs)
	}
	qty, err := conv.ParseDecimalField(in.Qty)
	if err != nil || !validComboDecimal(qty, false) || !qty.IsPositive() {
		return nil, errors.New("invalid combo quantity")
	}
	netPrice, err := conv.ParseDecimalField(in.NetPrice)
	if err != nil || !validComboDecimal(netPrice, true) {
		return nil, errors.New("invalid combo net price")
	}

	now := time.Now().Unix()
	seenContracts := make(map[int64]struct{}, len(in.Legs))
	legs := make([]*validatedComboLeg, 0, len(in.Legs))
	hasBuy := false
	hasSell := false
	ratios := make([]int64, 0, len(in.Legs))
	for _, inputLeg := range in.Legs {
		if inputLeg == nil || inputLeg.ContractId <= 0 {
			return nil, errors.New("each combo leg requires a contract")
		}
		if _, exists := seenContracts[inputLeg.ContractId]; exists {
			return nil, errors.New("combo legs must use distinct contracts")
		}
		seenContracts[inputLeg.ContractId] = struct{}{}
		if inputLeg.Side != common.Side_SIDE_BUY && inputLeg.Side != common.Side_SIDE_SELL {
			return nil, errors.New("each combo leg requires BUY or SELL")
		}
		if inputLeg.PositionEffect != option.PositionEffect_POSITION_EFFECT_OPEN {
			return nil, errors.New("combo V1 supports OPEN legs only")
		}
		if inputLeg.Ratio < 1 || inputLeg.Ratio > maxComboRatio {
			return nil, fmt.Errorf("each leg ratio must be between 1 and %d", maxComboRatio)
		}
		price, parseErr := conv.ParseDecimalField(inputLeg.Price)
		if parseErr != nil || !validComboDecimal(price, false) || !price.IsPositive() {
			return nil, errors.New("invalid combo leg price")
		}
		contract, findErr := svcCtx.OptionContractModel.FindOne(ctx, inputLeg.ContractId)
		if findErr != nil {
			if errors.Is(findErr, models.ErrNotFound) {
				return nil, errors.New("combo leg contract not found")
			}
			return nil, findErr
		}
		if contract.TenantId != tenantID ||
			contract.Status != int64(option.ContractStatus_CONTRACT_STATUS_TRADING) ||
			contract.IsDeleted == int64(common.YesNo_YES_NO_YES) ||
			now < contract.ListTime ||
			(contract.LastTradeTime <= 0 || now >= contract.LastTradeTime) {
			return nil, errors.New("combo leg contract is not tradable")
		}
		if contract.SettlementType != int64(option.SettlementType_SETTLEMENT_TYPE_CASH) {
			return nil, errors.New("combo V1 supports cash-settled options only")
		}
		if contract.SellerMarginMode != int64(option.SellerMarginMode_SELLER_MARGIN_MODE_ISOLATED) {
			return nil, errors.New("combo V1 requires isolated seller margin")
		}
		if contract.PriceTick.IsPositive() && !price.Mod(contract.PriceTick).IsZero() {
			return nil, errors.New("combo leg price does not follow price_tick")
		}
		legQty := qty.Mul(decimal.NewFromInt(inputLeg.Ratio))
		if !validComboDecimal(legQty, false) ||
			(contract.MinOrderQty.IsPositive() && legQty.LessThan(contract.MinOrderQty)) ||
			(contract.MaxOrderQty.IsPositive() && legQty.GreaterThan(contract.MaxOrderQty)) ||
			(contract.QtyStep.IsPositive() && !legQty.Mod(contract.QtyStep).IsZero()) {
			return nil, errors.New("combo leg quantity violates contract limits")
		}
		calendarDecision, calendarErr := logichelpers.IsContractTradingOpen(ctx, svcCtx, contract, now)
		if calendarErr != nil || calendarDecision == nil || !calendarDecision.Open {
			return nil, errors.New("combo leg trading calendar is closed")
		}
		leg := &validatedComboLeg{
			contract:   contract,
			side:       inputLeg.Side,
			ratio:      inputLeg.Ratio,
			price:      price,
			qty:        legQty,
			marginCoin: contract.SettleCoin,
		}
		switch inputLeg.Side {
		case common.Side_SIDE_BUY:
			hasBuy = true
			maxFeeRate := decimal.Max(contract.MakerFeeRate, contract.TakerFeeRate)
			leg.marginAmount = optionTurnover(contract, price, legQty).
				Mul(decimal.NewFromInt(1).Add(maxFeeRate)).
				Round(16)
		case common.Side_SIDE_SELL:
			hasSell = true
			market, marketErr := svcCtx.OptionMarketModel.FindOneByTenantIdContractId(
				ctx, tenantID, contract.Id,
			)
			if marketErr != nil || !logichelpers.IsUnderlyingFresh(market, now, 30) {
				return nil, errors.New("fresh underlying price is required for combo sell legs")
			}
			leg.marginAmount = optionSellerMargin(
				contract, market.UnderlyingPrice, price, legQty, false,
			)
		}
		if !validComboDecimal(leg.marginAmount, false) || !leg.marginAmount.IsPositive() {
			return nil, errors.New("invalid gross prefunding amount for combo leg")
		}
		legs = append(legs, leg)
		ratios = append(ratios, inputLeg.Ratio)
	}
	if !hasBuy || !hasSell {
		return nil, errors.New("combo order must contain at least one BUY and one SELL leg")
	}
	if gcdAll(ratios) != 1 {
		return nil, errors.New("combo leg ratios must be reduced to lowest terms")
	}

	sort.Slice(legs, func(i, j int) bool {
		return legs[i].contract.Id < legs[j].contract.Id
	})
	first := legs[0].contract
	calculatedNet := decimal.Zero
	strategyParts := make([]string, 0, len(legs))
	inverseParts := make([]string, 0, len(legs))
	for i, leg := range legs {
		contract := leg.contract
		if contract.UnderlyingSymbol != first.UnderlyingSymbol ||
			contract.ExpireTime != first.ExpireTime ||
			contract.SettleCoin != first.SettleCoin ||
			contract.QuoteCoin != first.QuoteCoin ||
			!contract.ContractUnit.Equal(first.ContractUnit) ||
			!optionMultiplier(contract).Equal(optionMultiplier(first)) {
			return nil, errors.New("combo legs must share underlying, expiry, coins, unit, and multiplier")
		}
		leg.legNo = int64(i + 1)
		sign := int64(1)
		if leg.side == common.Side_SIDE_SELL {
			sign = -1
		}
		signedRatio := sign * leg.ratio
		calculatedNet = calculatedNet.Add(
			leg.price.Mul(decimal.NewFromInt(signedRatio)),
		)
		strategyParts = append(strategyParts, fmt.Sprintf("%d:%+d", contract.Id, signedRatio))
		inverseParts = append(inverseParts, fmt.Sprintf("%d:%+d", contract.Id, -signedRatio))
	}
	if !calculatedNet.Equal(netPrice) {
		return nil, fmt.Errorf("net_price must equal signed leg-price sum %s", calculatedNet.String())
	}
	strategyText := strings.Join(strategyParts, "|")
	inverseStrategyText := strings.Join(inverseParts, "|")
	strategySum := sha256.Sum256([]byte(strategyText))
	inverseStrategySum := sha256.Sum256([]byte(inverseStrategyText))
	strategyKey := fmt.Sprintf("%x", strategySum[:])
	inverseStrategyKey := fmt.Sprintf("%x", inverseStrategySum[:])
	payloadHash, err := comboRequestPayloadHash(in)
	if err != nil {
		return nil, err
	}
	return &validatedComboOrder{
		qty:                qty,
		netPrice:           netPrice,
		strategyKey:        strategyKey,
		inverseStrategyKey: inverseStrategyKey,
		payloadHash:        payloadHash,
		legs:               legs,
	}, nil
}

func validComboDecimal(value decimal.Decimal, allowNegative bool) bool {
	if !allowNegative && value.IsNegative() {
		return false
	}
	if value.Exponent() < -16 {
		return false
	}
	limit := decimal.New(1, 16)
	return value.Abs().LessThan(limit)
}

func gcdAll(values []int64) int64 {
	result := int64(0)
	for _, value := range values {
		if value < 0 {
			value = -value
		}
		for value != 0 {
			result, value = value, result%value
		}
	}
	return result
}

func findComboOrderByNoOrID(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	tenantID, id int64,
	comboNo string,
) (*models.TOptionComboOrder, error) {
	if id > 0 {
		item, err := svcCtx.OptionComboOrderModel.FindOne(ctx, id)
		if err != nil {
			return nil, err
		}
		if item.TenantId != tenantID || (comboNo != "" && item.ComboNo != comboNo) {
			return nil, models.ErrNotFound
		}
		return item, nil
	}
	if comboNo == "" {
		return nil, models.ErrNotFound
	}
	return svcCtx.OptionComboOrderModel.FindOneByTenantIdComboNo(ctx, tenantID, comboNo)
}

func toComboOrderProto(item *models.TOptionComboOrder) *option.OptionComboOrder {
	if item == nil {
		return nil
	}
	return &option.OptionComboOrder{
		Id: item.Id, TenantId: item.TenantId, ComboNo: item.ComboNo,
		UserId: item.UserId, AccountId: item.AccountId,
		ClientComboId: item.ClientComboId, StrategyKey: item.StrategyKey,
		InverseStrategyKey: item.InverseStrategyKey,
		UnderlyingSymbol:   item.UnderlyingSymbol, ExpireTime: item.ExpireTime,
		SettleCoin: item.SettleCoin, QuoteCoin: item.QuoteCoin,
		OrderType: option.ComboOrderType(item.OrderType),
		NetPrice:  conv.FloatString(item.NetPrice), Qty: conv.FloatString(item.Qty),
		FilledQty:   conv.FloatString(item.FilledQty),
		UnfilledQty: conv.FloatString(item.UnfilledQty),
		Status:      option.ComboOrderStatus(item.Status), PayloadHash: item.PayloadHash,
		CancelReason: item.CancelReason, CancelTime: item.CancelTime,
		CreateTimes: item.CreateTimes, UpdateTimes: item.UpdateTimes,
	}
}

func toComboLegProto(item *models.TOptionComboOrderLeg) *option.OptionComboOrderLeg {
	if item == nil {
		return nil
	}
	return &option.OptionComboOrderLeg{
		Id: item.Id, TenantId: item.TenantId, ComboOrderId: item.ComboOrderId,
		LegNo: item.LegNo, ContractId: item.ContractId, Side: common.Side(item.Side),
		PositionEffect: option.PositionEffect(item.PositionEffect), Ratio: item.Ratio,
		Price: conv.FloatString(item.Price), Qty: conv.FloatString(item.Qty),
		FilledQty:    conv.FloatString(item.FilledQty),
		UnfilledQty:  conv.FloatString(item.UnfilledQty),
		ChildOrderId: item.ChildOrderId, CreateTimes: item.CreateTimes,
		UpdateTimes: item.UpdateTimes,
	}
}

func buildComboOrderDetail(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	item *models.TOptionComboOrder,
) (*option.OptionComboOrderDetail, error) {
	if item == nil {
		return nil, models.ErrNotFound
	}
	legs, err := svcCtx.OptionComboOrderLegModel.FindByComboOrderID(ctx, item.TenantId, item.Id)
	if err != nil {
		return nil, err
	}
	result := &option.OptionComboOrderDetail{
		ComboOrder: toComboOrderProto(item),
		Legs:       make([]*option.OptionComboOrderLeg, 0, len(legs)),
	}
	for _, leg := range legs {
		result.Legs = append(result.Legs, toComboLegProto(leg))
	}
	return result, nil
}
