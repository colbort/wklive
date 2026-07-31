package applogic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/option"
	logichelpers "wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetExerciseInstructionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetExerciseInstructionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetExerciseInstructionLogic {
	return &GetExerciseInstructionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取持仓当前生效的到期行权指令
func (l *GetExerciseInstructionLogic) GetExerciseInstruction(in *option.GetExerciseInstructionReq) (*option.GetExerciseInstructionResp, error) {
	tenantID, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	userID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	position, err := l.svcCtx.OptionPositionModel.FindOne(l.ctx, in.PositionId)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.GetExerciseInstructionResp{
				Base: helper.ErrResp(i18n.PositionNotFound, i18n.Translate(i18n.PositionNotFound, l.ctx)),
			}, nil
		}
		return nil, err
	}
	if position.TenantId != tenantID || position.UserId != userID ||
		position.AccountId != in.AccountId ||
		position.Side != int64(common.PositionSide_POSITION_SIDE_LONG) {
		return &option.GetExerciseInstructionResp{
			Base: helper.ErrResp(i18n.NoPermissionOperatePosition, i18n.Translate(i18n.NoPermissionOperatePosition, l.ctx)),
		}, nil
	}
	contract, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, position.ContractId)
	if err != nil {
		return nil, err
	}
	item, err := l.svcCtx.OptionExerciseInstructionModel.FindLatestByPosition(
		l.ctx, tenantID, position.Id,
	)
	if err == nil {
		return &option.GetExerciseInstructionResp{
			Base: helper.OkResp(), Data: logichelpers.ToExerciseInstructionProto(item),
		}, nil
	}
	if !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	return &option.GetExerciseInstructionResp{
		Base: helper.OkResp(),
		Data: &option.OptionExerciseInstruction{
			TenantId: tenantID, UserId: userID, AccountId: position.AccountId,
			ContractId: contract.Id, PositionId: position.Id,
			InstructionType: option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_AUTO,
			Status:          option.ExerciseInstructionStatus_EXERCISE_INSTRUCTION_STATUS_ACTIVE,
			CutoffTime:      contract.ExerciseCutoffTime,
		},
	}, nil
}
