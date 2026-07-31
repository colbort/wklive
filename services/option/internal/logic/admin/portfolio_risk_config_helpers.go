package adminlogic

import (
	"context"
	"errors"
	"regexp"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/option"
)

var errInvalidPortfolioRiskConfig = errors.New("invalid portfolio risk config")
var portfolioSettleCoinPattern = regexp.MustCompile(`^[A-Z0-9]{1,16}$`)

func portfolioConfigParamError(ctx context.Context) *option.GetPortfolioRiskConfigResp {
	return &option.GetPortfolioRiskConfigResp{
		Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, ctx)),
	}
}
