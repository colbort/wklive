// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_private

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/common/utils"
	"wklive/proto/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProfileLogic {
	return &UpdateProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateProfileLogic) UpdateProfile(req *types.UpdateProfileReq) (resp *types.UpdateProfileResp, err error) {
	userId, err := utils.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.UserCli.UpdateProfile(l.ctx, &user.UpdateProfileReq{
		UserId:      userId,
		Nickname:    req.Nickname,
		Avatar:      req.Avatar,
		Language:    req.Language,
		Timezone:    req.Timezone,
		Signature:   req.Signature,
		Gender:      user.Gender(req.Gender),
		Birthday:    req.Birthday,
		CountryCode: req.CountryCode,
		Province:    req.Province,
		City:        req.City,
		Address:     req.Address,
	})
	if err != nil {
		return nil, err
	}
	return &types.UpdateProfileResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
	}, nil
}
