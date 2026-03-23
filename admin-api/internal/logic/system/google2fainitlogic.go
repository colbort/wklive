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

type Google2FAInitLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGoogle2FAInitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Google2FAInitLogic {
	return &Google2FAInitLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *Google2FAInitLogic) Google2FAInit(req *types.Google2FAInitReq) (resp *types.Google2FAInitResp, err error) {
	result, err := l.svcCtx.SystemCli.Google2FAInit(l.ctx, &system.Google2FAInitReq{UserId: req.UserId})
	if err != nil {
		return nil, err
	}
	qrCode, _ := utils.GenerateGoogle2FAQRCodeDataURL(result.OtpauthUrl, 220)

	return &types.Google2FAInitResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Secret:     result.Secret,
		OtpauthUrl: result.OtpauthUrl,
		QrCode:     qrCode,
	}, nil
}
