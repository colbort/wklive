package adminlogic

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CreateSettlementPriceCorrectionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateSettlementPriceCorrectionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSettlementPriceCorrectionLogic {
	return &CreateSettlementPriceCorrectionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建人工结算价更正草案，必须由另一管理员确认
func (l *CreateSettlementPriceCorrectionLogic) CreateSettlementPriceCorrection(in *option.CreateSettlementPriceCorrectionReq) (*option.GetSettlementPriceResp, error) {
	operatorID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	price, err := conv.ParseDecimalField(in.DeliveryPrice)
	reason := strings.TrimSpace(in.Reason)
	var sourceIDs []string
	jsonErr := json.Unmarshal([]byte(in.SourceSnapshotIds), &sourceIDs)
	if in.TenantId <= 0 || in.ContractId <= 0 || operatorID <= 0 ||
		err != nil || !price.IsPositive() || reason == "" || len(reason) > 500 ||
		jsonErr != nil || len(sourceIDs) == 0 {
		return &option.GetSettlementPriceResp{
			Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx)),
		}, nil
	}
	for _, sourceID := range sourceIDs {
		if strings.TrimSpace(sourceID) == "" {
			return &option.GetSettlementPriceResp{
				Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx)),
			}, nil
		}
	}
	contract, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, in.ContractId)
	if err != nil {
		return nil, err
	}
	if contract.TenantId != in.TenantId {
		return &option.GetSettlementPriceResp{
			Base: helper.ErrResp(i18n.NoPermissionModify, i18n.Translate(i18n.NoPermissionModify, l.ctx)),
		}, nil
	}
	if _, err := l.svcCtx.OptionSettlementBatchModel.FindOneByTenantIdContractId(
		l.ctx, in.TenantId, in.ContractId,
	); err == nil {
		return &option.GetSettlementPriceResp{
			Base: helper.ErrResp(i18n.OperationNotAllowed, i18n.Translate(i18n.OperationNotAllowed, l.ctx)),
		}, nil
	} else if !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	now := time.Now().Unix()
	var created *models.TOptionSettlementPrice
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		priceModel := models.NewTOptionSettlementPriceModel(conn, l.svcCtx.Config.CacheRedis)
		latest, findErr := priceModel.FindLatestForUpdate(ctx, in.TenantId, in.ContractId)
		if findErr != nil && !errors.Is(findErr, models.ErrNotFound) {
			return findErr
		}
		if findErr == nil && latest.Status == int64(option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_PENDING) {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		}
		version := int64(1)
		if latest != nil {
			version = latest.Version + 1
		}
		confirmed, findErr := priceModel.FindLatestConfirmed(ctx, in.TenantId, in.ContractId)
		if findErr != nil && !errors.Is(findErr, models.ErrNotFound) {
			return findErr
		}
		supersedesID := int64(0)
		if confirmed != nil {
			supersedesID = confirmed.Id
		}
		created = &models.TOptionSettlementPrice{
			TenantId: in.TenantId, ContractId: in.ContractId,
			PriceSource: "manual-correction", WindowStart: contract.ExpireTime - contract.SettlementWindowSeconds,
			WindowEnd: contract.ExpireTime, SampleCount: int64(len(sourceIDs)),
			CalculationMethod: "MANUAL", DeliveryPrice: price,
			SourceSnapshotIds: in.SourceSnapshotIds, Version: version,
			Status:       int64(option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_PENDING),
			SupersedesId: supersedesID, ChangeReason: reason, CreatedBy: operatorID,
			CreateTimes: now, UpdateTimes: now,
		}
		result, insertErr := priceModel.Insert(ctx, created)
		if insertErr != nil {
			return insertErr
		}
		created.Id, insertErr = result.LastInsertId()
		return insertErr
	})
	if err != nil {
		return nil, err
	}
	return &option.GetSettlementPriceResp{
		Base: helper.OkResp(), Data: helpers.ToSettlementPriceProto(created),
	}, nil
}
