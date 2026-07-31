// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package option

import (
	"context"

	"wklive/app-api/internal/logicutil"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetExerciseInstructionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetExerciseInstructionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetExerciseInstructionLogic {
	return &SetExerciseInstructionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetExerciseInstructionLogic) SetExerciseInstruction(req *types.SetExerciseInstructionReq) (resp *types.GetExerciseInstructionResp, err error) {
	return logicutil.Proxy[types.GetExerciseInstructionResp](
		l.ctx, req, l.svcCtx.OptionCli.SetExerciseInstruction,
	)
}
