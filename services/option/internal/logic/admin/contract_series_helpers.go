package adminlogic

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"wklive/common/conv"
	"wklive/proto/common"
	"wklive/proto/option"
	logichelpers "wklive/services/option/internal/logic/helpers"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

const maxContractSeriesContracts = 500

var (
	contractSeriesCodePattern  = regexp.MustCompile(`^[A-Z0-9][A-Z0-9_-]{0,23}$`)
	contractSeriesKeyPattern   = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,95}$`)
	contractSeriesCyclePattern = regexp.MustCompile(
		`^[A-Z0-9][A-Z0-9_-]{0,31}$`,
	)
)

type preparedContractSeries struct {
	seriesCode       string
	requestKey       string
	referencePrice   decimal.Decimal
	referenceSource  string
	evidenceRef      string
	changeReason     string
	templateSnapshot string
	payloadHash      string
	template         *models.TOptionContract
	expiries         []*models.TOptionContractSeriesExpiry
	bands            []*models.TOptionContractSeriesStrikeBand
	strikes          []decimal.Decimal
	expectedCount    int64
}

type contractSeriesHashPayload struct {
	TenantID         int64                                     `json:"tenant_id"`
	RequestKey       string                                    `json:"request_key"`
	SeriesCode       string                                    `json:"series_code"`
	ReferencePrice   string                                    `json:"reference_price"`
	ReferenceSource  string                                    `json:"reference_source"`
	ReferenceTime    int64                                     `json:"reference_time"`
	EvidenceRef      string                                    `json:"evidence_ref"`
	ChangeReason     string                                    `json:"change_reason"`
	TemplateSnapshot string                                    `json:"template_snapshot"`
	Expiries         []*models.TOptionContractSeriesExpiry     `json:"expiries"`
	Bands            []*models.TOptionContractSeriesStrikeBand `json:"bands"`
}

func prepareContractSeries(in *option.CreateContractSeriesReq) (*preparedContractSeries, error) {
	if in == nil || in.TenantId <= 0 || in.ContractTemplate == nil {
		return nil, fmt.Errorf("missing contract series input")
	}
	requestKey := strings.TrimSpace(in.RequestKey)
	seriesCode := strings.ToUpper(strings.TrimSpace(in.SeriesCode))
	referenceSource := strings.TrimSpace(in.ReferenceSource)
	evidenceRef := strings.TrimSpace(in.EvidenceRef)
	changeReason := strings.TrimSpace(in.ChangeReason)
	if !contractSeriesKeyPattern.MatchString(requestKey) ||
		!contractSeriesCodePattern.MatchString(seriesCode) ||
		referenceSource == "" || len(referenceSource) > 128 ||
		evidenceRef == "" || len(evidenceRef) > 500 ||
		changeReason == "" || len(changeReason) > 500 ||
		in.ReferenceTime <= 0 || in.ReferenceTime > time.Now().Unix()+300 {
		return nil, fmt.Errorf("invalid contract series identity or evidence")
	}
	referencePrice, err := parseSeriesDecimal(in.ReferencePrice)
	if err != nil || !referencePrice.IsPositive() {
		return nil, fmt.Errorf("invalid reference price")
	}
	expiries, err := normalizeContractSeriesExpiries(in.TenantId, in.Expiries)
	if err != nil {
		return nil, err
	}
	bands, strikes, err := normalizeContractSeriesBands(in.TenantId, in.StrikeBands)
	if err != nil {
		return nil, err
	}
	expected := int64(len(expiries) * len(strikes) * 2)
	if expected < 2 || expected > maxContractSeriesContracts {
		return nil, fmt.Errorf("contract series size must be between 2 and %d", maxContractSeriesContracts)
	}
	template, err := contractSeriesTemplate(in.TenantId, in.ContractTemplate, expiries[0], referencePrice)
	if err != nil {
		return nil, err
	}
	snapshot, err := json.Marshal(template)
	if err != nil {
		return nil, err
	}
	payload := contractSeriesHashPayload{
		TenantID: in.TenantId, RequestKey: requestKey, SeriesCode: seriesCode,
		ReferencePrice: referencePrice.String(), ReferenceSource: referenceSource,
		ReferenceTime: in.ReferenceTime, EvidenceRef: evidenceRef, ChangeReason: changeReason,
		TemplateSnapshot: string(snapshot), Expiries: expiries, Bands: bands,
	}
	canonical, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(canonical)
	return &preparedContractSeries{
		seriesCode: seriesCode, requestKey: requestKey, referencePrice: referencePrice,
		referenceSource: referenceSource, evidenceRef: evidenceRef, changeReason: changeReason,
		templateSnapshot: string(snapshot), payloadHash: hex.EncodeToString(sum[:]), template: template,
		expiries: expiries, bands: bands, strikes: strikes, expectedCount: expected,
	}, nil
}

func normalizeContractSeriesExpiries(
	tenantID int64, inputs []*option.ContractSeriesExpiryInput,
) ([]*models.TOptionContractSeriesExpiry, error) {
	if len(inputs) == 0 || len(inputs) > 250 {
		return nil, fmt.Errorf("invalid expiry count")
	}
	items := make([]*models.TOptionContractSeriesExpiry, 0, len(inputs))
	seenSequence := make(map[int64]struct{}, len(inputs))
	seenExpiry := make(map[int64]struct{}, len(inputs))
	for _, input := range inputs {
		if input == nil {
			return nil, fmt.Errorf("nil expiry")
		}
		cycleCode := strings.ToUpper(strings.TrimSpace(input.CycleCode))
		lastTradeTime := input.LastTradeTime
		if lastTradeTime == 0 {
			lastTradeTime = input.ExerciseCutoffTime
		}
		if input.SequenceNo <= 0 || input.SequenceNo > 999 ||
			!contractSeriesCyclePattern.MatchString(cycleCode) ||
			input.ListTime <= 0 || lastTradeTime <= input.ListTime ||
			input.ExerciseCutoffTime < lastTradeTime ||
			input.ExpireTime < input.ExerciseCutoffTime || input.DeliverTime < input.ExpireTime {
			return nil, fmt.Errorf("invalid expiry specification")
		}
		if _, ok := seenSequence[input.SequenceNo]; ok {
			return nil, fmt.Errorf("duplicate expiry sequence")
		}
		if _, ok := seenExpiry[input.ExpireTime]; ok {
			return nil, fmt.Errorf("duplicate expiry time")
		}
		seenSequence[input.SequenceNo] = struct{}{}
		seenExpiry[input.ExpireTime] = struct{}{}
		items = append(items, &models.TOptionContractSeriesExpiry{
			TenantId: tenantID, SequenceNo: input.SequenceNo, CycleCode: cycleCode,
			ListTime: input.ListTime, LastTradeTime: lastTradeTime,
			ExerciseCutoffTime: input.ExerciseCutoffTime,
			ExpireTime:         input.ExpireTime, DeliverTime: input.DeliverTime,
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].SequenceNo < items[j].SequenceNo })
	return items, nil
}

func normalizeContractSeriesBands(
	tenantID int64, inputs []*option.ContractSeriesStrikeBandInput,
) ([]*models.TOptionContractSeriesStrikeBand, []decimal.Decimal, error) {
	if len(inputs) == 0 || len(inputs) > 100 {
		return nil, nil, fmt.Errorf("invalid strike band count")
	}
	items := make([]*models.TOptionContractSeriesStrikeBand, 0, len(inputs))
	seenSequence := make(map[int64]struct{}, len(inputs))
	for _, input := range inputs {
		if input == nil || input.SequenceNo <= 0 || input.SequenceNo > 999 {
			return nil, nil, fmt.Errorf("invalid strike band")
		}
		if _, ok := seenSequence[input.SequenceNo]; ok {
			return nil, nil, fmt.Errorf("duplicate strike band sequence")
		}
		lower, lowerErr := parseSeriesDecimal(input.LowerStrike)
		upper, upperErr := parseSeriesDecimal(input.UpperStrike)
		step, stepErr := parseSeriesDecimal(input.StrikeStep)
		if lowerErr != nil || upperErr != nil || stepErr != nil ||
			!lower.IsPositive() || upper.LessThan(lower) || !step.IsPositive() ||
			!upper.Sub(lower).Mod(step).IsZero() {
			return nil, nil, fmt.Errorf("invalid strike band range")
		}
		seenSequence[input.SequenceNo] = struct{}{}
		items = append(items, &models.TOptionContractSeriesStrikeBand{
			TenantId: tenantID, SequenceNo: input.SequenceNo,
			LowerStrike: lower, UpperStrike: upper, StrikeStep: step,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].LowerStrike.Equal(items[j].LowerStrike) {
			return items[i].SequenceNo < items[j].SequenceNo
		}
		return items[i].LowerStrike.LessThan(items[j].LowerStrike)
	})
	strikes := make([]decimal.Decimal, 0)
	for i, item := range items {
		if i > 0 && !item.LowerStrike.GreaterThan(items[i-1].UpperStrike) {
			return nil, nil, fmt.Errorf("overlapping strike bands")
		}
		for strike := item.LowerStrike; !strike.GreaterThan(item.UpperStrike); strike = strike.Add(item.StrikeStep) {
			if len(strikes) >= maxContractSeriesContracts/2 {
				return nil, nil, fmt.Errorf("too many strikes")
			}
			strikes = append(strikes, strike)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].SequenceNo < items[j].SequenceNo })
	return items, strikes, nil
}

func parseSeriesDecimal(value string) (decimal.Decimal, error) {
	parsed, err := decimal.NewFromString(strings.TrimSpace(value))
	if err != nil || parsed.Exponent() < -16 {
		return decimal.Zero, fmt.Errorf("decimal exceeds scale")
	}
	integerPart := parsed.Abs().Truncate(0).StringFixed(0)
	if len(integerPart) > 16 {
		return decimal.Zero, fmt.Errorf("decimal exceeds precision")
	}
	return parsed, nil
}

func contractSeriesTemplate(
	tenantID int64,
	in *option.CreateContractReq,
	expiry *models.TOptionContractSeriesExpiry,
	referencePrice decimal.Decimal,
) (*models.TOptionContract, error) {
	parse := func(value string) (decimal.Decimal, error) {
		return conv.ParseDecimalField(value)
	}
	contractUnit, err := parse(in.ContractUnit)
	if err != nil {
		return nil, err
	}
	minOrderQty, err := parse(in.MinOrderQty)
	if err != nil {
		return nil, err
	}
	maxOrderQty, err := parse(in.MaxOrderQty)
	if err != nil {
		return nil, err
	}
	priceTick, err := parse(in.PriceTick)
	if err != nil {
		return nil, err
	}
	qtyStep, err := parse(in.QtyStep)
	if err != nil {
		return nil, err
	}
	multiplier, err := parse(in.Multiplier)
	if err != nil {
		return nil, err
	}
	makerFeeRate, err := parseOptionalOptionRate(in.MakerFeeRate)
	if err != nil {
		return nil, err
	}
	takerFeeRate, err := parseOptionalOptionRate(in.TakerFeeRate)
	if err != nil {
		return nil, err
	}
	exerciseFeeRate, err := parseOptionalOptionRate(in.ExerciseFeeRate)
	if err != nil {
		return nil, err
	}
	autoExerciseThreshold, err := parse(in.AutoExerciseThreshold)
	if err != nil {
		return nil, err
	}
	maxUserLongQty, err := parse(in.MaxUserLongQty)
	if err != nil {
		return nil, err
	}
	maxUserShortQty, err := parse(in.MaxUserShortQty)
	if err != nil {
		return nil, err
	}
	maxOpenInterest, err := parse(in.MaxOpenInterest)
	if err != nil {
		return nil, err
	}
	orderPriceBandRatio, err := parseOptionalOptionRate(in.OrderPriceBandRatio)
	if err != nil {
		return nil, err
	}
	circuitBreakerRatio, err := parseOptionalOptionRate(in.CircuitBreakerRatio)
	if err != nil {
		return nil, err
	}
	initialMarginRate, err := parseOptionalOptionRate(in.InitialMarginRate)
	if err != nil {
		return nil, err
	}
	maintenanceMarginRate, err := parseOptionalOptionRate(in.MaintenanceMarginRate)
	if err != nil {
		return nil, err
	}
	minMarginRate, err := parseOptionalOptionRate(in.MinMarginRate)
	if err != nil {
		return nil, err
	}
	liquidationFeeRate, err := parseOptionalOptionRate(in.LiquidationFeeRate)
	if err != nil {
		return nil, err
	}
	sellerMarginMode := in.SellerMarginMode
	if sellerMarginMode == option.SellerMarginMode_SELLER_MARGIN_MODE_UNKNOWN {
		sellerMarginMode = option.SellerMarginMode_SELLER_MARGIN_MODE_DISABLED
	}
	deficitPolicy := in.LiquidationDeficitPolicy
	if deficitPolicy == option.LiquidationDeficitPolicy_LIQUIDATION_DEFICIT_POLICY_UNKNOWN {
		deficitPolicy = option.LiquidationDeficitPolicy_LIQUIDATION_DEFICIT_POLICY_MANUAL_REVIEW
	}
	calendarCode := strings.TrimSpace(in.TradingCalendarCode)
	if calendarCode == "" {
		calendarCode = logichelpers.DefaultTradingCalendarCode
	}
	item := &models.TOptionContract{
		TenantId: tenantID, ContractCode: "SERIES-TEMPLATE",
		UnderlyingSymbol: strings.TrimSpace(in.UnderlyingSymbol),
		UnderlyingCoin:   strings.TrimSpace(in.UnderlyingCoin),
		SettleCoin:       strings.TrimSpace(in.SettleCoin), QuoteCoin: strings.TrimSpace(in.QuoteCoin),
		OptionType:    int64(option.OptionType_OPTION_TYPE_CALL),
		ExerciseStyle: int64(in.ExerciseStyle), SettlementType: int64(in.SettlementType),
		StrikePrice: referencePrice, ContractUnit: contractUnit,
		MinOrderQty: minOrderQty, MaxOrderQty: maxOrderQty, PriceTick: priceTick,
		QtyStep: qtyStep, Multiplier: multiplier,
		ListTime: expiry.ListTime, LastTradeTime: expiry.LastTradeTime,
		ExerciseCutoffTime: expiry.ExerciseCutoffTime,
		ExpireTime:         expiry.ExpireTime, DeliverTime: expiry.DeliverTime,
		TradingCalendarCode: calendarCode, AutoExerciseThreshold: autoExerciseThreshold,
		MaxUserLongQty: maxUserLongQty, MaxUserShortQty: maxUserShortQty,
		MaxOpenInterest: maxOpenInterest, OrderPriceBandRatio: orderPriceBandRatio,
		CircuitBreakerRatio:     circuitBreakerRatio,
		GreeksMaxAgeSeconds:     in.GreeksMaxAgeSeconds,
		SettlementPriceSource:   strings.TrimSpace(in.SettlementPriceSource),
		SettlementPriceMethod:   strings.TrimSpace(in.SettlementPriceMethod),
		SettlementWindowSeconds: in.SettlementWindowSeconds,
		SettlementMinSamples:    in.SettlementMinSamples, IsAutoExercise: int64(in.IsAutoExercise),
		MakerFeeRate: makerFeeRate, TakerFeeRate: takerFeeRate, ExerciseFeeRate: exerciseFeeRate,
		FeeUserId: in.FeeUserId, FeeAccountId: in.FeeAccountId,
		SellerMarginMode: int64(sellerMarginMode), InitialMarginRate: initialMarginRate,
		MaintenanceMarginRate: maintenanceMarginRate, MinMarginRate: minMarginRate,
		LiquidationFeeRate: liquidationFeeRate, InsuranceUserId: in.InsuranceUserId,
		InsuranceAccountId: in.InsuranceAccountId, LiquidationDeficitPolicy: int64(deficitPolicy),
		PhysicalDeliveryPolicy:      int64(in.PhysicalDeliveryPolicy),
		PhysicalDeliveryCureSeconds: in.PhysicalDeliveryCureSeconds,
		Status:                      int64(option.ContractStatus_CONTRACT_STATUS_PENDING),
		Sort:                        int64(in.Sort), Remark: strings.TrimSpace(in.Remark),
		IsDeleted: int64(common.YesNo_YES_NO_NO),
	}
	if !validateSupportedContract(item) {
		return nil, fmt.Errorf("unsupported contract series template")
	}
	return item, nil
}

func decodeContractSeriesTemplate(snapshot string) (*models.TOptionContract, error) {
	var item models.TOptionContract
	if err := json.Unmarshal([]byte(snapshot), &item); err != nil {
		return nil, err
	}
	return &item, nil
}

func contractSeriesContractCode(
	seriesCode string, version, expirySequence int64, strikeIndex int, optionType option.OptionType,
) string {
	suffix := "C"
	if optionType == option.OptionType_OPTION_TYPE_PUT {
		suffix = "P"
	}
	return fmt.Sprintf("%s-V%d-E%03d-K%03d-%s", seriesCode, version, expirySequence, strikeIndex+1, suffix)
}

func cloneSeriesContract(
	template *models.TOptionContract,
	series *models.TOptionContractSeries,
	expiry *models.TOptionContractSeriesExpiry,
	strike decimal.Decimal,
	strikeIndex int,
	optionType option.OptionType,
	now int64,
) (*models.TOptionContract, error) {
	if template == nil || series == nil || expiry == nil {
		return nil, fmt.Errorf("missing generation input")
	}
	item := *template
	item.Id = 0
	item.TenantId = series.TenantId
	item.ContractCode = contractSeriesContractCode(series.SeriesCode, series.Version, expiry.SequenceNo, strikeIndex, optionType)
	if len(item.ContractCode) > 64 {
		return nil, fmt.Errorf("generated contract code is too long")
	}
	item.OptionType = int64(optionType)
	item.StrikePrice = strike
	item.ListTime = expiry.ListTime
	item.LastTradeTime = expiry.LastTradeTime
	item.ExerciseCutoffTime = expiry.ExerciseCutoffTime
	item.ExpireTime = expiry.ExpireTime
	item.DeliverTime = expiry.DeliverTime
	item.Status = int64(option.ContractStatus_CONTRACT_STATUS_PENDING)
	item.IsDeleted = int64(common.YesNo_YES_NO_NO)
	lineage := fmt.Sprintf("series=%d;version=%d", series.Id, series.Version)
	if item.Remark == "" {
		item.Remark = lineage
	} else {
		item.Remark = item.Remark + "; " + lineage
	}
	if len(item.Remark) > 500 {
		return nil, fmt.Errorf("generated contract remark is too long")
	}
	item.CreateTimes = now
	item.UpdateTimes = now
	if !validateSupportedContract(&item) {
		return nil, fmt.Errorf("generated contract failed contract validation")
	}
	return &item, nil
}

func toContractSeriesProto(
	item *models.TOptionContractSeries,
	expiries []*models.TOptionContractSeriesExpiry,
	bands []*models.TOptionContractSeriesStrikeBand,
) *option.OptionContractSeries {
	if item == nil {
		return nil
	}
	result := &option.OptionContractSeries{
		Id: item.Id, TenantId: item.TenantId, RequestKey: item.RequestKey,
		SeriesCode: item.SeriesCode, Version: item.Version, SupersedesId: item.SupersedesId,
		Status: option.ContractSeriesStatus(item.Status), TemplateContractId: item.TemplateContractId,
		UnderlyingSymbol: item.UnderlyingSymbol, ReferencePrice: item.ReferencePrice.String(),
		ReferenceSource: item.ReferenceSource, ReferenceTime: item.ReferenceTime,
		EvidenceRef: item.EvidenceRef, ChangeReason: item.ChangeReason, PayloadHash: item.PayloadHash,
		ExpectedContractCount: item.ExpectedContractCount, GeneratedContractCount: item.GeneratedContractCount,
		CreatedBy: item.CreatedBy, ReviewedBy: item.ReviewedBy, ReviewReason: item.ReviewReason,
		ReviewedAt: item.ReviewedAt, GeneratedAt: item.GeneratedAt,
		LaunchStatus:     option.ContractSeriesLaunchStatus(item.LaunchStatus),
		LaunchReviewedBy: item.LaunchReviewedBy, LaunchReviewReason: item.LaunchReviewReason,
		LaunchReviewedAt: item.LaunchReviewedAt,
		CreateTimes:      item.CreateTimes, UpdateTimes: item.UpdateTimes,
	}
	for _, expiry := range expiries {
		result.Expiries = append(result.Expiries, &option.OptionContractSeriesExpiry{
			Id: expiry.Id, TenantId: expiry.TenantId, SeriesId: expiry.SeriesId,
			SequenceNo: expiry.SequenceNo, CycleCode: expiry.CycleCode,
			ListTime: expiry.ListTime, LastTradeTime: expiry.LastTradeTime,
			ExerciseCutoffTime: expiry.ExerciseCutoffTime,
			ExpireTime:         expiry.ExpireTime, DeliverTime: expiry.DeliverTime, CreateTimes: expiry.CreateTimes,
		})
	}
	for _, band := range bands {
		result.StrikeBands = append(result.StrikeBands, &option.OptionContractSeriesStrikeBand{
			Id: band.Id, TenantId: band.TenantId, SeriesId: band.SeriesId,
			SequenceNo: band.SequenceNo, LowerStrike: band.LowerStrike.String(),
			UpperStrike: band.UpperStrike.String(), StrikeStep: band.StrikeStep.String(),
			CreateTimes: band.CreateTimes,
		})
	}
	return result
}

func toContractSeriesDetailProto(item *models.TOptionContractSeriesDetail) *option.OptionContractSeriesDetail {
	if item == nil {
		return nil
	}
	return &option.OptionContractSeriesDetail{
		Id: item.Id, TenantId: item.TenantId, SeriesId: item.SeriesId, ExpiryId: item.ExpiryId,
		OptionType: option.OptionType(item.OptionType), StrikePrice: item.StrikePrice.String(),
		ContractCode: item.ContractCode, ContractId: item.ContractId, CreateTimes: item.CreateTimes,
	}
}
