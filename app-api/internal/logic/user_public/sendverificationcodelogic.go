// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_public

import (
	"context"
	"strings"

	"wklive/app-api/internal/logicutil"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/system"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendVerificationCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendVerificationCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendVerificationCodeLogic {
	return &SendVerificationCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendVerificationCodeLogic) SendVerificationCode(req *types.SendVerificationCodeReq) (resp *types.RespBase, err error) {
	tenantCode, err := utils.GetTenantCodeFromCtx(l.ctx)
	tenantCode = strings.TrimSpace(tenantCode)
	if err != nil || tenantCode == "" {
		return &types.RespBase{
			Code: i18n.InvalidRequest,
			Msg:  i18n.Translate(i18n.InvalidRequest, l.ctx),
		}, nil
	}

	tenant, err := l.svcCtx.SystemCli.SysTenantDetail(l.ctx, &system.SysTenantDetailReq{
		TenantCode: &tenantCode,
	})
	if err != nil {
		return logicutil.SystemErrorResp[types.RespBase](l.ctx, err)
	}
	if tenant.GetBase().GetCode() != helper.OkResp().Code {
		return &types.RespBase{
			Code: tenant.GetBase().GetCode(),
			Msg:  tenant.GetBase().GetMsg(),
		}, nil
	}

	return logicutil.Proxy[types.RespBase](l.ctx, &system.SendVerificationCodeReq{
		TenantId: tenant.GetData().GetId(),
		Channel:  system.VerificationCodeChannel(req.Channel),
		Email:    req.Email,
		Phone:    req.Phone,
		Scene:    system.VerificationCodeScene(req.Scene),
	}, l.svcCtx.SystemCli.SendVerificationCode)
}
