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

type TestVerificationCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTestVerificationCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TestVerificationCodeLogic {
	return &TestVerificationCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TestVerificationCodeLogic) TestVerificationCode(req *types.TestVerificationCodeReq) (resp *types.RespBase, err error) {
	return logicutil.Proxy[types.RespBase](l.ctx, req, l.svcCtx.SystemCli.TestVerificationCode)
}
