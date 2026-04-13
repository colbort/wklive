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

type GetIdentityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetIdentityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetIdentityLogic {
	return &GetIdentityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetIdentityLogic) GetIdentity() (resp *types.GetIdentityResp, err error) {
	userId, err := utils.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.UserCli.GetIdentity(l.ctx, &user.GetIdentityReq{
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}
	return &types.GetIdentityResp{
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
			VerifyBy:      result.Identity.VerifyBy,
			CreateTimes:   result.Identity.CreateTimes,
			UpdateTimes:   result.Identity.UpdateTimes,
		},
	}, nil
}
