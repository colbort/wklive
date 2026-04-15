package user

import (
	"context"
	"strings"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckUserReferrerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCheckUserReferrerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckUserReferrerLogic {
	return &CheckUserReferrerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CheckUserReferrerLogic) CheckUserReferrer(req *types.CheckUserReferrerReq) (resp *types.CheckUserReferrerResp, err error) {
	inviteCode := strings.TrimSpace(req.InviteCode)
	referrer, err := resolveReferrerByInviteCode(l.svcCtx, l.ctx, 0, inviteCode)
	if err != nil {
		return nil, err
	}

	if referrer == nil {
		return &types.CheckUserReferrerResp{
			RespBase: types.RespBase{
				Code: 404,
				Msg:  "推荐人不存在",
			},
			Exists: false,
		}, nil
	}

	resp = &types.CheckUserReferrerResp{
		RespBase: types.RespBase{
			Code: 200,
			Msg:  "success",
		},
		Exists: true,
	}
	resp.Data = types.UserReferrerInfo{
		UserId:     referrer.Id,
		Username:   referrer.Username,
		Nickname:   referrer.Nickname,
		InviteCode: referrer.InviteCode,
	}

	return resp, nil
}
