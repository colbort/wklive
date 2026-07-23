package itickadminlogic

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"
	"wklive/services/itick/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreatePriceFormulaLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreatePriceFormulaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePriceFormulaLogic {
	return &CreatePriceFormulaLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Price Engine 公式内容不可原地修改，变更必须创建新版本。
func (l *CreatePriceFormulaLogic) CreatePriceFormula(in *itick.CreatePriceFormulaReq) (*itick.PriceFormulaResp, error) {
	components, err := normalizePriceFormulaReq(in)
	if err != nil {
		return nil, err
	}
	outputAuthority, err := l.svcCtx.AuthorityRegistryModel.FindEnabled(l.ctx, in.Authority)
	if err != nil || outputAuthority == nil || !outputAuthority.Allows(in.SnapshotKind) {
		return nil, errors.New("output authority is not enabled for snapshot kind")
	}
	for _, component := range components {
		authority, findErr := l.svcCtx.AuthorityRegistryModel.FindEnabled(l.ctx, component.Authority)
		if findErr != nil || authority == nil || !authority.Allows(component.Kind) {
			return nil, errors.New("component authority is not enabled for snapshot kind")
		}
	}
	raw, err := json.Marshal(components)
	if err != nil {
		return nil, err
	}
	now := utils.NowMillis()
	row := &models.TItickPriceFormula{FormulaNo: strings.TrimSpace(in.FormulaNo), Authority: in.Authority, SnapshotKind: in.SnapshotKind, CategoryCode: in.CategoryCode, Market: in.Market, Symbol: in.Symbol, Algorithm: int64(in.Algorithm), FormulaVersion: strings.TrimSpace(in.FormulaVersion), Components: string(raw), MaxLookbackMs: in.MaxLookbackMs, MaxDeviationBps: in.MaxDeviationBps, IntervalMs: in.IntervalMs, Status: 2, CreateTimes: now, UpdateTimes: now}
	result, err := l.svcCtx.PriceFormulaModel.Insert(l.ctx, row)
	if err != nil {
		return nil, err
	}
	row.Id, _ = result.LastInsertId()
	if in.Activate {
		if err = l.svcCtx.PriceFormulaModel.ActivateVersion(l.ctx, row.Id, now); err != nil {
			return nil, err
		}
		row, err = l.svcCtx.PriceFormulaModel.FindOne(l.ctx, row.Id)
		if err != nil {
			return nil, err
		}
	}
	return &itick.PriceFormulaResp{Base: helper.OkResp(), Data: toPriceFormulaProto(row)}, nil
}
