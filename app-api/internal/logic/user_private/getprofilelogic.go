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

type GetProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProfileLogic {
	return &GetProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProfileLogic) GetProfile() (resp *types.GetProfileResp, err error) {
	userId, err := utils.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.UserCli.GetProfile(l.ctx, &user.GetProfileReq{
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}
	return &types.GetProfileResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: types.UserProfile{
			Identity: types.UserIdentity{
				Id:            result.Profile.Identity.Id,
				TenantId:      result.Profile.Identity.TenantId,
				UserId:        result.Profile.Identity.UserId,
				Phone:         result.Profile.Identity.Phone,
				Email:         result.Profile.Identity.Email,
				RealName:      result.Profile.Identity.RealName,
				Gender:        int64(result.Profile.Identity.Gender.Number()),
				Birthday:      result.Profile.Identity.Birthday,
				CountryCode:   result.Profile.Identity.CountryCode,
				Province:      result.Profile.Identity.Province,
				City:          result.Profile.Identity.City,
				Address:       result.Profile.Identity.Address,
				IdType:        int64(result.Profile.Identity.IdType.Number()),
				IdNo:          result.Profile.Identity.IdNo,
				FrontImage:    result.Profile.Identity.FrontImage,
				BackImage:     result.Profile.Identity.BackImage,
				HandheldImage: result.Profile.Identity.HandheldImage,
				KycLevel:      int64(result.Profile.Identity.KycLevel.Number()),
				VerifyStatus:  int64(result.Profile.Identity.VerifyStatus.Number()),
				RejectReason:  result.Profile.Identity.RejectReason,
				SubmitTime:    result.Profile.Identity.SubmitTime,
				VerifyTime:    result.Profile.Identity.VerifyTime,
				VerifyBy:      result.Profile.Identity.VerifyBy,
				CreateTimes:    result.Profile.Identity.CreateTimes,
				UpdateTimes:    result.Profile.Identity.UpdateTimes,
			},
			Security: types.UserSecurity{
				Id:              result.Profile.Security.Id,
				TenantId:        result.Profile.Security.TenantId,
				UserId:          result.Profile.Security.UserId,
				HasPayPassword:  result.Profile.Security.HasPayPassword,
				GoogleEnabled:   result.Profile.Security.GoogleEnabled,
				LoginErrorCount: result.Profile.Security.LoginErrorCount,
				PayErrorCount:   result.Profile.Security.PayErrorCount,
				LockUntil:       result.Profile.Security.LockUntil,
				RiskLevel:       int64(result.Profile.Security.RiskLevel.Number()),
				CreateTimes:      result.Profile.Security.CreateTimes,
				UpdateTimes:      result.Profile.Security.UpdateTimes,
			},
		},
	}, nil

}
