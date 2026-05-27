// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_private

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/proto/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateIdentityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateIdentityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateIdentityLogic {
	return &UpdateIdentityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateIdentityLogic) UpdateIdentity(req *types.UpdateIdentityReq) (resp *types.UpdateIdentityResp, err error) {
	result, err := l.svcCtx.UserCli.UpdateIdentity(l.ctx, &user.UpdateIdentityReq{
		Phone:       req.Phone,
		Email:       req.Email,
		RealName:    req.RealName,
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
	return &types.UpdateIdentityResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: types.UserIdentity{
			Id:            result.Identity.Id,
			TenantId:      result.Identity.TenantId,
			UserId:        result.Identity.UserId,
			Phone:         result.Identity.Phone,
			Email:         result.Identity.Email,
			RealName:      result.Identity.RealName,
			Gender:        int64(result.Identity.Gender),
			Birthday:      result.Identity.Birthday,
			CountryCode:   result.Identity.CountryCode,
			Province:      result.Identity.Province,
			City:          result.Identity.City,
			Address:       result.Identity.Address,
			IdType:        int64(result.Identity.IdType),
			IdNo:          result.Identity.IdNo,
			FrontImage:    result.Identity.FrontImage,
			BackImage:     result.Identity.BackImage,
			HandheldImage: result.Identity.HandheldImage,
			KycLevel:      int64(result.Identity.KycLevel),
			VerifyStatus:  int64(result.Identity.VerifyStatus),
			RejectReason:  result.Identity.RejectReason,
			SubmitTime:    result.Identity.SubmitTime,
			VerifyTime:    result.Identity.VerifyTime,
		},
	}, nil
}
