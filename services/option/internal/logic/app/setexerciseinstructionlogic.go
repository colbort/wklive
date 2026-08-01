package applogic

import (
	"context"
	"errors"
	"strings"
	"time"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/option"
	logichelpers "wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type SetExerciseInstructionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetExerciseInstructionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetExerciseInstructionLogic {
	return &SetExerciseInstructionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置到期行权指令；截止时间前可用新幂等号产生新版本
func (l *SetExerciseInstructionLogic) SetExerciseInstruction(in *option.SetExerciseInstructionReq) (*option.GetExerciseInstructionResp, error) {
	tenantID, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	userID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	clientID := strings.TrimSpace(in.ClientInstructionId)
	if clientID == "" || len(clientID) > 64 || !validExerciseInstructionType(in.InstructionType) {
		return exerciseInstructionParamError(l.ctx), nil
	}
	if existing, findErr := l.svcCtx.OptionExerciseInstructionModel.
		FindOneByTenantIdUserIdClientInstructionId(l.ctx, tenantID, userID, clientID); findErr == nil {
		return exerciseInstructionReplayResponse(l.ctx, existing, in), nil
	} else if !errors.Is(findErr, models.ErrNotFound) {
		return nil, findErr
	}

	position, contract, response, err := l.loadExerciseInstructionScope(
		tenantID, userID, in.AccountId, in.PositionId, in.ContractId,
	)
	if err != nil || response != nil {
		return response, err
	}
	now := time.Now().Unix()
	if !allowsExerciseSubmission(contract.Status) || now < contract.ListTime ||
		contract.ExerciseCutoffTime <= 0 || now >= contract.ExerciseCutoffTime {
		return &option.GetExerciseInstructionResp{
			Base: helper.ErrResp(i18n.OperationNotAllowed, i18n.Translate(i18n.OperationNotAllowed, l.ctx)),
		}, nil
	}
	if blocked, checkErr := l.svcCtx.OptionCorporateActionContractModel.IsContractMigrationActive(
		l.ctx, contract.TenantId, contract.Id,
	); checkErr != nil {
		return nil, checkErr
	} else if blocked {
		return &option.GetExerciseInstructionResp{
			Base: helper.ErrResp(i18n.OperationNotAllowed, i18n.Translate(i18n.OperationNotAllowed, l.ctx)),
		}, nil
	}

	var result *models.TOptionExerciseInstruction
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		contractModel := models.NewTOptionContractModel(conn, l.svcCtx.Config.CacheRedis)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)
		instructionModel := models.NewTOptionExerciseInstructionModel(conn, l.svcCtx.Config.CacheRedis)
		actionContractModel := models.NewTOptionCorporateActionContractModel(conn, l.svcCtx.Config.CacheRedis)
		currentContract, lockErr := contractModel.FindOneForUpdate(ctx, contract.Id)
		if lockErr != nil {
			return lockErr
		}
		txNow := time.Now().Unix()
		if currentContract.TenantId != tenantID ||
			currentContract.SettlementType != int64(option.SettlementType_SETTLEMENT_TYPE_CASH) ||
			currentContract.IsAutoExercise != int64(common.YesNo_YES_NO_YES) ||
			!allowsExerciseSubmission(currentContract.Status) ||
			txNow < currentContract.ListTime || currentContract.ExerciseCutoffTime <= 0 ||
			txNow >= currentContract.ExerciseCutoffTime {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		}
		currentPosition, lockErr := positionModel.FindOneForUpdate(ctx, position.Id)
		if lockErr != nil {
			return lockErr
		}
		if currentPosition.TenantId != tenantID || currentPosition.UserId != userID ||
			currentPosition.AccountId != in.AccountId ||
			currentPosition.ContractId != contract.Id ||
			currentPosition.Side != int64(common.PositionSide_POSITION_SIDE_LONG) ||
			currentPosition.Status != int64(option.PositionStatus_POSITION_STATUS_HOLDING) ||
			!currentPosition.PositionQty.IsPositive() {
			return i18n.StatusError(ctx, i18n.NoPermissionOperatePosition)
		}
		blocked, checkErr := actionContractModel.IsContractMigrationActive(
			ctx, currentPosition.TenantId, currentPosition.ContractId,
		)
		if checkErr != nil {
			return checkErr
		}
		if blocked {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		}
		if replay, findErr := instructionModel.FindOneByTenantIdUserIdClientInstructionId(
			ctx, tenantID, userID, clientID,
		); findErr == nil {
			if !sameExerciseInstruction(replay, in) {
				return i18n.StatusError(ctx, i18n.ParamError)
			}
			result = replay
			return nil
		} else if !errors.Is(findErr, models.ErrNotFound) {
			return findErr
		}

		version := int64(1)
		supersedesID := int64(0)
		latest, findErr := instructionModel.FindLatestByPositionForUpdate(ctx, tenantID, position.Id)
		if findErr == nil {
			version = latest.Version + 1
			supersedesID = latest.Id
			latest.Status = int64(option.ExerciseInstructionStatus_EXERCISE_INSTRUCTION_STATUS_SUPERSEDED)
			latest.UpdateTimes = now
			if updateErr := instructionModel.Update(ctx, latest); updateErr != nil {
				return updateErr
			}
		} else if !errors.Is(findErr, models.ErrNotFound) {
			return findErr
		}
		result = &models.TOptionExerciseInstruction{
			TenantId: tenantID, UserId: userID, AccountId: in.AccountId,
			ContractId: contract.Id, PositionId: position.Id,
			ClientInstructionId: clientID, InstructionType: int64(in.InstructionType),
			Version:      version,
			Status:       int64(option.ExerciseInstructionStatus_EXERCISE_INSTRUCTION_STATUS_ACTIVE),
			SupersedesId: supersedesID, CutoffTime: currentContract.ExerciseCutoffTime,
			CreateTimes: now, UpdateTimes: now,
		}
		insertResult, insertErr := instructionModel.Insert(ctx, result)
		if insertErr != nil {
			return insertErr
		}
		result.Id, insertErr = insertResult.LastInsertId()
		return insertErr
	})
	if err != nil {
		if existing, findErr := l.svcCtx.OptionExerciseInstructionModel.
			FindOneByTenantIdUserIdClientInstructionId(l.ctx, tenantID, userID, clientID); findErr == nil {
			return exerciseInstructionReplayResponse(l.ctx, existing, in), nil
		}
		return nil, err
	}
	return &option.GetExerciseInstructionResp{
		Base: helper.OkResp(), Data: logichelpers.ToExerciseInstructionProto(result),
	}, nil
}

func validExerciseInstructionType(value option.ExerciseInstructionType) bool {
	switch value {
	case option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_AUTO,
		option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_DO_NOT_EXERCISE,
		option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_EXERCISE:
		return true
	default:
		return false
	}
}

func sameExerciseInstruction(
	item *models.TOptionExerciseInstruction,
	in *option.SetExerciseInstructionReq,
) bool {
	return item != nil && in != nil &&
		item.AccountId == in.AccountId &&
		item.PositionId == in.PositionId &&
		(in.ContractId == 0 || item.ContractId == in.ContractId) &&
		item.InstructionType == int64(in.InstructionType)
}

func exerciseInstructionReplayResponse(
	ctx context.Context,
	item *models.TOptionExerciseInstruction,
	in *option.SetExerciseInstructionReq,
) *option.GetExerciseInstructionResp {
	if !sameExerciseInstruction(item, in) {
		return exerciseInstructionParamError(ctx)
	}
	return &option.GetExerciseInstructionResp{
		Base: helper.OkResp(), Data: logichelpers.ToExerciseInstructionProto(item),
	}
}

func exerciseInstructionParamError(ctx context.Context) *option.GetExerciseInstructionResp {
	return &option.GetExerciseInstructionResp{
		Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, ctx)),
	}
}

func (l *SetExerciseInstructionLogic) loadExerciseInstructionScope(
	tenantID, userID, accountID, positionID, contractID int64,
) (
	*models.TOptionPosition,
	*models.TOptionContract,
	*option.GetExerciseInstructionResp,
	error,
) {
	position, err := l.svcCtx.OptionPositionModel.FindOne(l.ctx, positionID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, nil, &option.GetExerciseInstructionResp{
				Base: helper.ErrResp(i18n.PositionNotFound, i18n.Translate(i18n.PositionNotFound, l.ctx)),
			}, nil
		}
		return nil, nil, nil, err
	}
	if position.TenantId != tenantID || position.UserId != userID ||
		position.AccountId != accountID ||
		(contractID != 0 && position.ContractId != contractID) ||
		position.Side != int64(common.PositionSide_POSITION_SIDE_LONG) ||
		position.Status != int64(option.PositionStatus_POSITION_STATUS_HOLDING) {
		return nil, nil, &option.GetExerciseInstructionResp{
			Base: helper.ErrResp(i18n.NoPermissionOperatePosition, i18n.Translate(i18n.NoPermissionOperatePosition, l.ctx)),
		}, nil
	}
	contract, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, position.ContractId)
	if err != nil {
		return nil, nil, nil, err
	}
	if contract.SettlementType != int64(option.SettlementType_SETTLEMENT_TYPE_CASH) ||
		contract.IsAutoExercise != int64(common.YesNo_YES_NO_YES) ||
		contract.ExerciseCutoffTime <= 0 {
		return nil, nil, &option.GetExerciseInstructionResp{
			Base: helper.ErrResp(i18n.OperationNotAllowed, i18n.Translate(i18n.OperationNotAllowed, l.ctx)),
		}, nil
	}
	return position, contract, nil, nil
}
