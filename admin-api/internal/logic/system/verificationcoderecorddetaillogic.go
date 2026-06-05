// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type VerificationCodeRecordDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVerificationCodeRecordDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerificationCodeRecordDetailLogic {
	return &VerificationCodeRecordDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VerificationCodeRecordDetailLogic) VerificationCodeRecordDetail(req *types.VerificationCodeRecordDetailReq) (resp *types.VerificationCodeRecordDetailResp, err error) {
	return logicutil.Proxy[types.VerificationCodeRecordDetailResp](l.ctx, req, l.svcCtx.SystemCli.VerificationCodeRecordDetail)
}
