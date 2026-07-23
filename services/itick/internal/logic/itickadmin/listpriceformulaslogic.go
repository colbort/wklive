package itickadminlogic

import (
	"context"
	"errors"
	"strings"

	"wklive/common/pageutil"
	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"
	"wklive/services/itick/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPriceFormulasLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPriceFormulasLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPriceFormulasLogic {
	return &ListPriceFormulasLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListPriceFormulasLogic) ListPriceFormulas(in *itick.ListPriceFormulasReq) (*itick.ListPriceFormulasResp, error) {
	if in == nil || in.Page == nil {
		return nil, errors.New("page is required")
	}
	rows, count, err := l.svcCtx.PriceFormulaModel.FindPage(l.ctx, models.PriceFormulaFilter{Authority: strings.ToLower(strings.TrimSpace(in.Authority)), SnapshotKind: strings.ToUpper(strings.TrimSpace(in.SnapshotKind)), CategoryCode: strings.ToLower(strings.TrimSpace(in.CategoryCode)), Market: strings.ToUpper(strings.TrimSpace(in.Market)), Symbol: strings.ToUpper(strings.TrimSpace(in.Symbol)), Status: int64(in.Status)}, in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}
	data := make([]*itick.PriceFormulaData, 0, len(rows))
	for _, row := range rows {
		data = append(data, toPriceFormulaProto(row))
	}
	lastID := int64(0)
	if len(rows) > 0 {
		lastID = rows[len(rows)-1].Id
	}
	return &itick.ListPriceFormulasResp{Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(rows), count, lastID), Data: data}, nil
}
