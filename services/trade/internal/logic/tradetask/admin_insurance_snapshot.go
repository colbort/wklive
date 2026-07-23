package tradetasklogic

import (
	"context"
	"errors"
	"strings"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"
)

type AdminInsuranceSnapshotLogic struct {
	ctx context.Context
	svc *svc.ServiceContext
}

func NewAdminInsuranceSnapshotLogic(ctx context.Context, s *svc.ServiceContext) *AdminInsuranceSnapshotLogic {
	return &AdminInsuranceSnapshotLogic{ctx, s}
}

func (l *AdminInsuranceSnapshotLogic) SetInsuranceFundAccount(in *trade.SetInsuranceFundAccountReq) (*trade.CommonResp, error) {
	tenant := adminTenantID(l.ctx, in.TenantId)
	assetCode := strings.ToUpper(strings.TrimSpace(in.SettleAsset))
	if tenant <= 0 || assetCode == "" {
		return &trade.CommonResp{Base: helper.ErrResp(i18n.ParamError, "invalid insurance fund account")}, nil
	}
	if in.SymbolId > 0 {
		s, err := l.svc.TradeSymbolModel.FindOne(l.ctx, in.SymbolId)
		if err != nil || s.TenantId != tenant {
			return &trade.CommonResp{Base: helper.ErrResp(i18n.BusinessDataNotFound, "symbol not found")}, nil
		}
	}
	now := utils.NowMillis()
	status := int64(in.Status)
	if status == 0 {
		status = 1
	}
	adl := int64(in.AdlEnabled)
	if adl == 0 {
		adl = 2
	}
	row := &models.TContractInsuranceFundAccount{TenantId: tenant, SymbolId: in.SymbolId, SettleAsset: assetCode, AdlEnabled: adl, Status: status, Version: in.Version, CreateTimes: now, UpdateTimes: now}
	var err error
	if in.Id > 0 {
		old, e := l.svc.ContractInsuranceFundModel.FindOne(l.ctx, in.Id)
		if errors.Is(e, models.ErrNotFound) || (e == nil && old.TenantId != tenant) {
			return &trade.CommonResp{Base: helper.ErrResp(i18n.BusinessDataNotFound, "insurance fund account not found")}, nil
		}
		if e != nil {
			return nil, e
		}
		if old.Version != in.Version {
			return &trade.CommonResp{Base: helper.ErrResp(i18n.ParamError, "insurance fund account version conflict")}, nil
		}
		row.Id, row.CreateTimes = old.Id, old.CreateTimes
		err = l.svc.ContractInsuranceFundModel.Update(l.ctx, row)
	} else {
		_, err = l.svc.ContractInsuranceFundModel.Insert(l.ctx, row)
	}
	if err != nil {
		return nil, err
	}
	return &trade.CommonResp{Base: helper.OkResp()}, nil
}
func (l *AdminInsuranceSnapshotLogic) GetInsuranceFundAccountList(in *trade.GetInsuranceFundAccountListReq) (*trade.GetInsuranceFundAccountListResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	rows, total, err := l.svc.ContractInsuranceFundModel.FindPage(l.ctx, adminTenantID(l.ctx, in.TenantId), in.SymbolId, int64(in.Status), cursor, limit, strings.ToUpper(strings.TrimSpace(in.SettleAsset)))
	if err != nil {
		return nil, err
	}
	resp := &trade.GetInsuranceFundAccountListResp{}
	last := int64(0)
	for _, v := range rows {
		resp.Data = append(resp.Data, &trade.InsuranceFundAccount{Id: v.Id, TenantId: v.TenantId, SymbolId: v.SymbolId, SettleAsset: v.SettleAsset, AdlEnabled: common.YesNo(v.AdlEnabled), Status: common.Enable(v.Status), Version: v.Version, CreateTimes: v.CreateTimes, UpdateTimes: v.UpdateTimes})
		last = v.Id
	}
	resp.Base = pageutil.Base(cursor, limit, len(rows), total, last)
	return resp, nil
}
func (l *AdminInsuranceSnapshotLogic) GetMarketSnapshotList(in *trade.GetMarketSnapshotListReq) (*trade.GetMarketSnapshotListResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	rows, total, err := l.svc.TradeMarketSnapshotModel.FindPage(l.ctx, adminTenantID(l.ctx, in.TenantId), in.SymbolId, cursor, limit, in.StartTime, in.EndTime, strings.ToUpper(strings.TrimSpace(in.SnapshotKind)))
	if err != nil {
		return nil, err
	}
	resp := &trade.GetMarketSnapshotListResp{}
	last := int64(0)
	for _, v := range rows {
		resp.Data = append(resp.Data, &trade.TradeMarketSnapshot{Id: v.Id, SnapshotId: v.SnapshotId, SnapshotKind: v.SnapshotKind, SymbolId: v.SymbolId, Source: v.Source, Price: v.Price.String(), MarkPrice: v.MarkPrice.String(), IndexPrice: v.IndexPrice.String(), FundingRate: v.FundingRate.String(), SourceTimestamp: v.SourceTimestamp, SnapshotTimestamp: v.SnapshotTimestamp, Revision: v.Revision, FormulaVersion: v.FormulaVersion, Confirmed: common.YesNo(v.Confirmed), RawPayload: v.RawPayload, CreateTimes: v.CreateTimes})
		last = v.Id
	}
	resp.Base = pageutil.Base(cursor, limit, len(rows), total, last)
	return resp, nil
}
