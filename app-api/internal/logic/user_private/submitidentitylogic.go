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

type SubmitIdentityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubmitIdentityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitIdentityLogic {
	return &SubmitIdentityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitIdentityLogic) SubmitIdentity(req *types.SubmitIdentityReq) (resp *types.SubmitIdentityResp, err error) {
	userId, err := utils.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.UserCli.SubmitIdentity(l.ctx, &user.SubmitIdentityReq{
		UserId:        userId,
		Phone:         req.Phone,
		Email:         req.Email,
		RealName:      req.RealName,
		Gender:        user.Gender(req.Gender),
		Birthday:      req.Birthday,
		CountryCode:   req.CountryCode,
		Province:      req.Province,
		City:          req.City,
		Address:       req.Address,
		IdType:        user.IdType(req.IdType),
		IdNo:          req.IdNo,
		FrontImage:    req.FrontImage,
		BackImage:     req.BackImage,
		HandheldImage: req.HandheldImage,
		KycLevel:      user.KycLevel(req.KycLevel),
	})
	if err != nil {
		return nil, err
	}
	return &types.SubmitIdentityResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
	}, nil
}
