package marketlogic

import (
	"context"
	"errors"
	"strings"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/market"
	"wklive/services/market/internal/logic/helpers"
	"wklive/services/market/internal/pkg/utils"
	"wklive/services/market/internal/svc"
	"wklive/services/market/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResolveTenantProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResolveTenantProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResolveTenantProductLogic {
	return &ResolveTenantProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ResolveTenantProduct returns only a currently usable tenant product. It is
// intentionally stricter than the admin detail API, which must also show
// disabled records for maintenance.
func (l *ResolveTenantProductLogic) ResolveTenantProduct(in *market.ResolveTenantProductReq) (*market.ResolveTenantProductResp, error) {
	if in.GetTenantId() <= 0 || in.GetTenantProductId() <= 0 {
		return l.notFound(), nil
	}

	item, err := l.svcCtx.MarketTenantProductModel.FindOne(l.ctx, in.GetTenantProductId())
	if errors.Is(err, models.ErrNotFound) {
		return l.notFound(), nil
	}
	if err != nil {
		return nil, err
	}
	if item.TenantId != in.GetTenantId() || item.Enabled != 1 || item.AppVisible != 1 {
		return l.notFound(), nil
	}

	product, err := l.svcCtx.MarketProductModel.FindOne(l.ctx, item.ProductId)
	if errors.Is(err, models.ErrNotFound) {
		return l.notFound(), nil
	}
	if err != nil {
		return nil, err
	}
	if product.Enabled != 1 || product.AppVisible != 1 {
		return l.notFound(), nil
	}
	if product.CategoryType == int64(market.CategoryType_CATEGORY_TYPE_STOCK) &&
		!utils.StockExchangeMatchesMarket(product.Market, product.Exchange) {
		l.Errorf("tenant stock product market mismatch tenant_product_id=%d product_id=%d market=%s exchange=%s symbol=%s",
			item.Id, product.Id, product.Market, product.Exchange, product.Symbol)
		return &market.ResolveTenantProductResp{
			Base: helper.ErrResp(i18n.ParamError, "stock product market and exchange do not match"),
		}, nil
	}

	data := helpers.ToTenantProductProto(item, product)
	data.Market = strings.ToUpper(strings.TrimSpace(data.Market))
	data.Symbol = strings.ToUpper(strings.TrimSpace(data.Symbol))
	data.BaseCoin = strings.ToUpper(strings.TrimSpace(data.BaseCoin))
	data.QuoteCoin = strings.ToUpper(strings.TrimSpace(data.QuoteCoin))
	return &market.ResolveTenantProductResp{Base: helper.OkResp(), Data: data}, nil
}

func (l *ResolveTenantProductLogic) notFound() *market.ResolveTenantProductResp {
	return &market.ResolveTenantProductResp{
		Base: helper.ErrResp(i18n.BusinessDataNotFound, "enabled tenant product not found"),
	}
}
