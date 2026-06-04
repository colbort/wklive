// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"strings"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserBaseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserBaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserBaseLogic {
	return &UpdateUserBaseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserBaseLogic) UpdateUserBase(req *types.UpdateUserBaseReq) (resp *types.UpdateUserBaseResp, err error) {
	referrerUserId := req.ReferrerUserId
	if strings.TrimSpace(req.ReferrerInviteCode) != "" {
		referrer, err := resolveReferrerByInviteCode(l.svcCtx, l.ctx, req.TenantId, req.ReferrerInviteCode)
		if err != nil {
			return nil, err
		}
		if referrer == nil {
			return &types.UpdateUserBaseResp{
				RespBase: types.RespBase{
					Code: 404,
					Msg:  "推荐人不存在",
				},
			}, nil
		}
		referrerUserId = referrer.Id
	}

	protoReq := &user.UpdateUserBaseReq{
		TenantId:       req.TenantId,
		UserId:         req.UserId,
		Username:       req.Username,
		Nickname:       req.Nickname,
		Avatar:         req.Avatar,
		Language:       req.Language,
		Timezone:       req.Timezone,
		Signature:      req.Signature,
		Source:         req.Source,
		ReferrerUserId: referrerUserId,
		Remark:         req.Remark,
		Phone:          req.Phone,
		Email:          req.Email,
	}
	return logicutil.Proxy[types.UpdateUserBaseResp](l.ctx, protoReq, l.svcCtx.UserCli.UpdateUserBase)
}
