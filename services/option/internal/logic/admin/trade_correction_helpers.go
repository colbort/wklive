package adminlogic

import (
	"errors"
	"fmt"
	"strings"

	"wklive/common/conv"
	"wklive/proto/option"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

var (
	errActiveTradeCorrection  = errors.New("active trade correction already exists")
	errInvalidTradeCorrection = errors.New("invalid trade correction")
)

type validatedTradeCorrectionLeg struct {
	UserID    int64
	AccountID int64
	Coin      string
	Direction option.TradeCorrectionLegDirection
	Amount    decimal.Decimal
}

func validateTradeCorrectionLegs(
	inputs []*option.TradeCorrectionLegInput,
	trade *models.TOptionTrade,
	contract *models.TOptionContract,
) ([]validatedTradeCorrectionLeg, error) {
	if trade == nil || contract == nil || len(inputs) < 2 || len(inputs) > 20 ||
		trade.FeeCoin == "" {
		return nil, errInvalidTradeCorrection
	}
	type participant struct {
		userID    int64
		accountID int64
	}
	allowed := map[participant]struct{}{
		{trade.BuyUserId, trade.BuyAccountId}:   {},
		{trade.SellUserId, trade.SellAccountId}: {},
	}
	if contract.FeeUserId > 0 {
		allowed[participant{contract.FeeUserId, contract.FeeAccountId}] = struct{}{}
	}
	if contract.InsuranceUserId > 0 {
		allowed[participant{contract.InsuranceUserId, contract.InsuranceAccountId}] = struct{}{}
	}

	debit, credit := decimal.Zero, decimal.Zero
	result := make([]validatedTradeCorrectionLeg, 0, len(inputs))
	for _, input := range inputs {
		if input == nil || input.UserId <= 0 ||
			strings.TrimSpace(input.Coin) != trade.FeeCoin {
			return nil, errInvalidTradeCorrection
		}
		if _, ok := allowed[participant{input.UserId, input.AccountId}]; !ok {
			return nil, fmt.Errorf("%w: account is not a trade or approved platform participant", errInvalidTradeCorrection)
		}
		amount, err := conv.ParseDecimalField(input.Amount)
		if err != nil || !amount.IsPositive() || amount.Exponent() < -16 {
			return nil, errInvalidTradeCorrection
		}
		switch input.Direction {
		case option.TradeCorrectionLegDirection_TRADE_CORRECTION_LEG_DIRECTION_DEBIT:
			debit = debit.Add(amount)
		case option.TradeCorrectionLegDirection_TRADE_CORRECTION_LEG_DIRECTION_CREDIT:
			credit = credit.Add(amount)
		default:
			return nil, errInvalidTradeCorrection
		}
		result = append(result, validatedTradeCorrectionLeg{
			UserID: input.UserId, AccountID: input.AccountId,
			Coin: trade.FeeCoin, Direction: input.Direction, Amount: amount,
		})
	}
	if !debit.IsPositive() || !credit.IsPositive() || !debit.Equal(credit) {
		return nil, fmt.Errorf("%w: debit and credit must balance exactly", errInvalidTradeCorrection)
	}
	return result, nil
}
