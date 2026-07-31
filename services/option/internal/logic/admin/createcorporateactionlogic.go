package adminlogic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"wklive/common/generate"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/option"
	logichelpers "wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CreateCorporateActionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCorporateActionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCorporateActionLogic {
	return &CreateCorporateActionLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

type preparedCorporateActionContract struct {
	input       *option.CorporateActionContractInput
	numerator   decimal.Decimal
	denominator decimal.Decimal
	haltNo      string
}

// 登记不可覆盖的公司行动版本；所有受影响合约立即停牌并撤单
func (l *CreateCorporateActionLogic) CreateCorporateAction(
	in *option.CreateCorporateActionReq,
) (*option.GetCorporateActionResp, error) {
	operatorID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return corporateActionError(l.ctx, i18n.PermissionDenied), nil
	}
	prepared, err := l.validateAndPrepare(in, operatorID)
	if err != nil {
		return corporateActionError(l.ctx, i18n.ParamError), nil
	}

	now := time.Now().Unix()
	var created *models.TOptionCorporateAction
	var createdContracts []*models.TOptionCorporateActionContract
	var createdHalts []*models.TOptionTradingHalt
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		actionModel := models.NewTOptionCorporateActionModel(conn, l.svcCtx.Config.CacheRedis)
		actionContractModel := models.NewTOptionCorporateActionContractModel(conn, l.svcCtx.Config.CacheRedis)
		contractModel := models.NewTOptionContractModel(conn, l.svcCtx.Config.CacheRedis)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)
		haltModel := models.NewTOptionTradingHaltModel(conn, l.svcCtx.Config.CacheRedis)
		eventModel := models.NewTOptionTradingControlEventModel(conn, l.svcCtx.Config.CacheRedis)

		version := int64(1)
		latest, findErr := actionModel.FindLatestVersionForUpdate(ctx, in.TenantId, strings.TrimSpace(in.ExternalEventRef))
		if findErr == nil {
			if latest.Status == int64(option.CorporateActionStatus_CORPORATE_ACTION_STATUS_DRAFT) ||
				latest.Status == int64(option.CorporateActionStatus_CORPORATE_ACTION_STATUS_APPROVED) ||
				latest.Status == int64(option.CorporateActionStatus_CORPORATE_ACTION_STATUS_EXECUTING) {
				return i18n.StatusError(ctx, i18n.OperationNotAllowed)
			}
			version = latest.Version + 1
		} else if !errors.Is(findErr, models.ErrNotFound) {
			return findErr
		}
		if existing, findErr := actionModel.FindOneByTenantIdEventNo(
			ctx, in.TenantId, strings.TrimSpace(in.EventNo),
		); findErr == nil && existing != nil {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		} else if findErr != nil && !errors.Is(findErr, models.ErrNotFound) {
			return findErr
		}

		created = &models.TOptionCorporateAction{
			TenantId: in.TenantId, EventNo: strings.TrimSpace(in.EventNo),
			ExternalEventRef: strings.TrimSpace(in.ExternalEventRef), Version: version,
			UnderlyingSymbol: strings.ToUpper(strings.TrimSpace(in.UnderlyingSymbol)),
			ActionType:       int64(in.ActionType),
			Status:           int64(option.CorporateActionStatus_CORPORATE_ACTION_STATUS_DRAFT),
			AnnouncementTime: in.AnnouncementTime, ExTime: in.ExTime, RecordTime: in.RecordTime,
			EffectiveTime: in.EffectiveTime, PayTime: in.PayTime,
			EvidenceRef: strings.TrimSpace(in.EvidenceRef), EvidenceHash: strings.TrimSpace(in.EvidenceHash),
			Description: strings.TrimSpace(in.Description), CreatedBy: operatorID,
			CreateTimes: now, UpdateTimes: now,
		}
		result, insertErr := actionModel.Insert(ctx, created)
		if insertErr != nil {
			return insertErr
		}
		created.Id, insertErr = result.LastInsertId()
		if insertErr != nil {
			return insertErr
		}

		seenSources := make(map[int64]struct{}, len(prepared))
		for _, item := range prepared {
			source, lockErr := contractModel.FindOneForUpdate(ctx, item.input.SourceContractId)
			if lockErr != nil {
				return lockErr
			}
			if _, duplicate := seenSources[source.Id]; duplicate {
				return i18n.StatusError(ctx, i18n.OperationNotAllowed)
			}
			seenSources[source.Id] = struct{}{}
			if source.TenantId != in.TenantId ||
				!strings.EqualFold(source.UnderlyingSymbol, created.UnderlyingSymbol) ||
				source.Status != int64(option.ContractStatus_CONTRACT_STATUS_TRADING) {
				return i18n.StatusError(ctx, i18n.OperationNotAllowed)
			}
			if _, findErr = haltModel.FindActiveByContract(ctx, source.TenantId, source.Id); findErr == nil {
				return i18n.StatusError(ctx, i18n.OperationNotAllowed)
			} else if !errors.Is(findErr, models.ErrNotFound) {
				return findErr
			}
			if validateErr := validateCorporateActionSuccessor(
				ctx, contractModel, positionModel, source, item.input,
			); validateErr != nil {
				return validateErr
			}

			mapping := &models.TOptionCorporateActionContract{
				TenantId: in.TenantId, ActionId: created.Id,
				SourceContractId: source.Id, SuccessorContractId: item.input.SuccessorContractId,
				ExecutionMode:     int64(item.input.ExecutionMode),
				QuantityNumerator: item.numerator, QuantityDenominator: item.denominator,
				Status:      int64(option.CorporateActionContractStatus_CORPORATE_ACTION_CONTRACT_STATUS_HALTED),
				CreateTimes: now, UpdateTimes: now,
			}
			result, insertErr = actionContractModel.Insert(ctx, mapping)
			if insertErr != nil {
				return insertErr
			}
			mapping.Id, insertErr = result.LastInsertId()
			if insertErr != nil {
				return insertErr
			}

			halt := &models.TOptionTradingHalt{
				TenantId: in.TenantId, HaltNo: item.haltNo,
				ActiveKey: fmt.Sprintf("CONTRACT:%d", source.Id), ContractId: source.Id,
				Source:      int64(option.TradingHaltSource_TRADING_HALT_SOURCE_CORPORATE_ACTION),
				Status:      int64(option.TradingHaltStatus_TRADING_HALT_STATUS_ACTIVE),
				Reason:      fmt.Sprintf("corporate action %s", created.EventNo),
				EvidenceRef: created.EvidenceRef, StartedAt: now, CreatedBy: operatorID,
				CreateTimes: now, UpdateTimes: now,
			}
			result, insertErr = haltModel.Insert(ctx, halt)
			if insertErr != nil {
				return insertErr
			}
			halt.Id, insertErr = result.LastInsertId()
			if insertErr != nil {
				return insertErr
			}
			mapping.HaltId = halt.Id
			if updateErr := actionContractModel.Update(ctx, mapping); updateErr != nil {
				return updateErr
			}
			source.Status = int64(option.ContractStatus_CONTRACT_STATUS_PAUSED)
			source.UpdateTimes = now
			if updateErr := contractModel.Update(ctx, source); updateErr != nil {
				return updateErr
			}
			_, insertErr = eventModel.Insert(ctx, &models.TOptionTradingControlEvent{
				TenantId: source.TenantId, ContractId: source.Id, EventType: "CORPORATE_ACTION_HALTED",
				Reason:     created.EventNo,
				Detail:     fmt.Sprintf("actionId=%d mappingId=%d haltNo=%s", created.Id, mapping.Id, halt.HaltNo),
				OperatorId: operatorID, CreateTimes: now,
			})
			if insertErr != nil {
				return insertErr
			}
			createdContracts = append(createdContracts, mapping)
			createdHalts = append(createdHalts, halt)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	cancelLogic := NewForceCancelContractOrdersLogic(l.ctx, l.svcCtx)
	for index, mapping := range createdContracts {
		source, findErr := l.svcCtx.OptionContractModel.FindOne(l.ctx, mapping.SourceContractId)
		if findErr != nil {
			return nil, findErr
		}
		total, success, failed, cancelErr := cancelLogic.forceCancelAll(
			source, "CORPORATE_ACTION:"+created.EventNo, true,
		)
		lastError := ""
		if cancelErr != nil {
			lastError = cancelErr.Error()
			if len(lastError) > 500 {
				lastError = lastError[:500]
			}
		}
		halt := createdHalts[index]
		if updateErr := l.updateCorporateActionHalt(halt.Id, total, success, failed, lastError); updateErr != nil {
			return nil, updateErr
		}
	}
	return &option.GetCorporateActionResp{
		Base: helper.OkResp(), Data: logichelpers.ToCorporateActionProto(created, createdContracts),
	}, nil
}

func (l *CreateCorporateActionLogic) validateAndPrepare(
	in *option.CreateCorporateActionReq, operatorID int64,
) ([]preparedCorporateActionContract, error) {
	if in == nil || in.TenantId <= 0 || operatorID <= 0 ||
		strings.TrimSpace(in.EventNo) == "" || len(strings.TrimSpace(in.EventNo)) > 96 ||
		strings.TrimSpace(in.ExternalEventRef) == "" || len(strings.TrimSpace(in.ExternalEventRef)) > 128 ||
		strings.TrimSpace(in.UnderlyingSymbol) == "" || len(strings.TrimSpace(in.UnderlyingSymbol)) > 32 ||
		in.ActionType == option.CorporateActionType_CORPORATE_ACTION_TYPE_UNKNOWN ||
		in.AnnouncementTime <= 0 || in.EffectiveTime <= 0 ||
		strings.TrimSpace(in.EvidenceRef) == "" || len(strings.TrimSpace(in.EvidenceRef)) > 500 ||
		strings.TrimSpace(in.EvidenceHash) == "" || len(strings.TrimSpace(in.EvidenceHash)) > 128 ||
		strings.TrimSpace(in.Description) == "" || len(strings.TrimSpace(in.Description)) > 1000 ||
		len(in.Contracts) == 0 || len(in.Contracts) > 1000 {
		return nil, logichelpers.ErrCorporateActionInexact
	}
	result := make([]preparedCorporateActionContract, 0, len(in.Contracts))
	for _, contract := range in.Contracts {
		if contract == nil || contract.SourceContractId <= 0 ||
			(contract.ExecutionMode != option.CorporateActionExecutionMode_CORPORATE_ACTION_EXECUTION_MODE_AUTO_CASH_SUCCESSOR &&
				contract.ExecutionMode != option.CorporateActionExecutionMode_CORPORATE_ACTION_EXECUTION_MODE_MANUAL_ONLY) {
			return nil, logichelpers.ErrCorporateActionInexact
		}
		numerator, err := logichelpers.ParsePositiveCorporateActionInteger(contract.QuantityNumerator)
		if err != nil {
			return nil, err
		}
		denominator, err := logichelpers.ParsePositiveCorporateActionInteger(contract.QuantityDenominator)
		if err != nil {
			return nil, err
		}
		if contract.ExecutionMode == option.CorporateActionExecutionMode_CORPORATE_ACTION_EXECUTION_MODE_AUTO_CASH_SUCCESSOR &&
			in.ActionType != option.CorporateActionType_CORPORATE_ACTION_TYPE_SPLIT &&
			in.ActionType != option.CorporateActionType_CORPORATE_ACTION_TYPE_REVERSE_SPLIT {
			return nil, logichelpers.ErrCorporateActionInexact
		}
		haltNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "option_halt_id", "OH", "")
		if err != nil {
			return nil, err
		}
		result = append(result, preparedCorporateActionContract{
			input: contract, numerator: numerator, denominator: denominator, haltNo: haltNo,
		})
	}
	return result, nil
}

func validateCorporateActionSuccessor(
	ctx context.Context,
	contractModel models.TOptionContractModel,
	positionModel models.TOptionPositionModel,
	source *models.TOptionContract,
	input *option.CorporateActionContractInput,
) error {
	if input.ExecutionMode == option.CorporateActionExecutionMode_CORPORATE_ACTION_EXECUTION_MODE_MANUAL_ONLY {
		return nil
	}
	successor, err := contractModel.FindOneForUpdate(ctx, input.SuccessorContractId)
	if err != nil {
		return err
	}
	if successor.Id == source.Id || successor.TenantId != source.TenantId ||
		successor.Status != int64(option.ContractStatus_CONTRACT_STATUS_PENDING) ||
		source.SettlementType != int64(option.SettlementType_SETTLEMENT_TYPE_CASH) ||
		successor.SettlementType != int64(option.SettlementType_SETTLEMENT_TYPE_CASH) ||
		!strings.EqualFold(source.UnderlyingSymbol, successor.UnderlyingSymbol) ||
		source.OptionType != successor.OptionType || source.ExerciseStyle != successor.ExerciseStyle ||
		source.SettleCoin != successor.SettleCoin || source.QuoteCoin != successor.QuoteCoin ||
		source.ExpireTime != successor.ExpireTime {
		return i18n.StatusError(ctx, i18n.OperationNotAllowed)
	}
	total, err := positionModel.CountHoldingByContract(ctx, successor.TenantId, successor.Id)
	if err != nil {
		return err
	}
	if total != 0 {
		return i18n.StatusError(ctx, i18n.OperationNotAllowed)
	}
	return nil
}

func (l *CreateCorporateActionLogic) updateCorporateActionHalt(
	haltID, total, success, failed int64, lastError string,
) error {
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		haltModel := models.NewTOptionTradingHaltModel(
			sqlx.NewSqlConnFromSession(session), l.svcCtx.Config.CacheRedis,
		)
		halt, err := haltModel.FindOneForUpdate(ctx, haltID)
		if err != nil {
			return err
		}
		halt.CancelTotal += total
		halt.CancelSuccess += success
		halt.CancelFailed += failed
		halt.LastErrorMsg = lastError
		halt.UpdateTimes = time.Now().Unix()
		return haltModel.Update(ctx, halt)
	})
}

func corporateActionError(ctx context.Context, code int32) *option.GetCorporateActionResp {
	return &option.GetCorporateActionResp{
		Base: helper.ErrResp(code, i18n.Translate(code, ctx)),
	}
}
