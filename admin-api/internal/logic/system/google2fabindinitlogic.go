// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/common/utils"
	"wklive/proto/system"

	"github.com/zeromicro/go-zero/core/logx"
)

type Google2FABindInitLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGoogle2FABindInitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Google2FABindInitLogic {
	return &Google2FABindInitLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *Google2FABindInitLogic) Google2FABindInit(req *types.Google2FABindInitReq) (resp *types.Google2FABindInitResp, err error) {
	result, err := l.svcCtx.SystemCli.Google2FAInit(l.ctx, &system.Google2FAInitReq{UserId: req.UserId})
	if err != nil {
		return nil, err
	}
	qrCode, _ := utils.GenerateGoogle2FAQRCodeDataURL(result.OtpauthUrl, 220)

	return &types.Google2FABindInitResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Secret:     result.Secret,
		OtpauthUrl: result.OtpauthUrl,
		QrCode:     qrCode,
	}, nil
}
