package adminlogic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/market"
	"wklive/services/market/internal/svc"
	"wklive/services/market/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetAuthorityRegistryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetAuthorityRegistryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetAuthorityRegistryLogic {
	return &SetAuthorityRegistryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Authority 名称和 producer_type 创建后不可修改；数据库写入由 model 完成。
func (l *SetAuthorityRegistryLogic) SetAuthorityRegistry(in *market.SetAuthorityRegistryReq) (*market.AuthorityRegistryResp, error) {
	if in == nil {
		return nil, errors.New("authority registry request is required")
	}
	authority, providerCode, producerType, kinds, rawKinds, err := normalizeAuthorityRegistry(
		in.Authority,
		in.ProviderCode,
		in.ProducerType,
		in.AllowedKinds,
	)
	if err != nil {
		return nil, err
	}
	status := int64(in.Status)
	if status != int64(common.Enable_ENABLE_ENABLED) &&
		status != int64(common.Enable_ENABLE_DISABLED) {
		return nil, errors.New("authority status must be enabled or disabled")
	}
	now := utils.NowMillis()
	if in.Id == 0 {
		if in.Version != 0 {
			return nil, errors.New("new authority version must be zero")
		}
		if _, findErr := l.svcCtx.AuthorityRegistryAdminModel.FindOneByAuthority(l.ctx, authority); findErr == nil {
			return nil, errors.New("authority already exists")
		} else if !errors.Is(findErr, models.ErrNotFound) {
			return nil, findErr
		}
		row := &models.TItickAuthorityRegistry{
			Authority:    authority,
			ProviderCode: providerCode,
			ProducerType: producerType,
			AllowedKinds: rawKinds,
			Status:       status,
			Version:      0,
			CreateTimes:  now,
			UpdateTimes:  now,
		}
		row.Id, err = l.svcCtx.AuthorityRegistryAdminModel.Create(l.ctx, row)
		if err != nil {
			return nil, err
		}
		data, err := authorityRegistryProto(row)
		if err != nil {
			return nil, err
		}
		return &market.AuthorityRegistryResp{Base: helper.OkResp(), Data: data}, nil
	}

	row, err := l.svcCtx.AuthorityRegistryAdminModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if row.Authority != authority ||
		row.ProviderCode != providerCode ||
		row.ProducerType != producerType {
		return nil, errors.New("authority, provider_code and producer_type are immutable")
	}
	if row.Version != in.Version {
		return nil, errors.New("authority registry version conflict")
	}
	removesKind, err := removesAuthorityKind(row.AllowedKinds, kinds)
	if err != nil {
		return nil, err
	}
	if status == int64(common.Enable_ENABLE_DISABLED) || removesKind {
		references, countErr := l.svcCtx.AuthorityRegistryAdminModel.CountActiveFormulaReferences(l.ctx, authority)
		if countErr != nil {
			return nil, countErr
		}
		if references > 0 {
			return nil, errors.New("authority is referenced by active price formulas")
		}
	}
	updated, err := l.svcCtx.AuthorityRegistryAdminModel.UpdateConfigVersioned(
		l.ctx,
		row.Id,
		row.Version,
		rawKinds,
		status,
		now,
	)
	if err != nil {
		return nil, err
	}
	if !updated {
		return nil, errors.New("authority registry version conflict")
	}
	row.AllowedKinds = rawKinds
	row.Status = status
	row.Version++
	row.UpdateTimes = now
	data, err := authorityRegistryProto(row)
	if err != nil {
		return nil, err
	}
	return &market.AuthorityRegistryResp{Base: helper.OkResp(), Data: data}, nil
}
