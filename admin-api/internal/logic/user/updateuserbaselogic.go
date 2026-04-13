// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

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
	result, err := l.svcCtx.UserCli.UpdateUserBase(l.ctx, &user.UpdateUserBaseReq{
		TenantId:       req.TenantId,
		UserId:         req.UserId,
		Username:       req.Username,
		Nickname:       req.Nickname,
		Avatar:         req.Avatar,
		Language:       req.Language,
		Timezone:       req.Timezone,
		Signature:      req.Signature,
		Source:         req.Source,
		ReferrerUserId: req.ReferrerUserId,
		Remark:         req.Remark,
		Phone:          req.Phone,
		Email:          req.Email,
	})
	if err != nil {
		return nil, err
	}

	// Map UserDetail (similar to GetUserDetail)
	detail := types.UserDetail{
		Base: types.UserBase{
			Id:             result.Detail.Base.Id,
			TenantId:       result.Detail.Base.TenantId,
			UserNo:         result.Detail.Base.UserNo,
			Username:       result.Detail.Base.Username,
			Nickname:       result.Detail.Base.Nickname,
			Avatar:         result.Detail.Base.Avatar,
			PasswordHash:   result.Detail.Base.PasswordHash,
			RegisterType:   int64(result.Detail.Base.RegisterType),
			Status:         int64(result.Detail.Base.Status),
			MemberLevel:    result.Detail.Base.MemberLevel,
			Language:       result.Detail.Base.Language,
			Timezone:       result.Detail.Base.Timezone,
			InviteCode:     result.Detail.Base.InviteCode,
			Signature:      result.Detail.Base.Signature,
			Source:         result.Detail.Base.Source,
			ReferrerUserId: result.Detail.Base.ReferrerUserId,
			RegisterTime:   result.Detail.Base.RegisterTime,
			RegisterIp:     result.Detail.Base.RegisterIp,
			LastLoginTime:  result.Detail.Base.LastLoginTime,
			LastLoginIp:    result.Detail.Base.LastLoginIp,
			IsGuest:        result.Detail.Base.IsGuest,
			IsRecharge:     result.Detail.Base.IsRecharge,
			DeviceId:       result.Detail.Base.DeviceId,
			Fingerprint:    result.Detail.Base.Fingerprint,
			Remark:         result.Detail.Base.Remark,
			Deleted:        result.Detail.Base.Deleted,
			CreateTimes:    result.Detail.Base.CreateTimes,
			UpdateTimes:    result.Detail.Base.UpdateTimes,
		},
		Identity: types.UserIdentity{
			Id:            result.Detail.Identity.Id,
			TenantId:      result.Detail.Identity.TenantId,
			UserId:        result.Detail.Identity.UserId,
			Phone:         result.Detail.Identity.Phone,
			Email:         result.Detail.Identity.Email,
			RealName:      result.Detail.Identity.RealName,
			Gender:        int64(result.Detail.Identity.Gender),
			Birthday:      result.Detail.Identity.Birthday,
			CountryCode:   result.Detail.Identity.CountryCode,
			Province:      result.Detail.Identity.Province,
			City:          result.Detail.Identity.City,
			Address:       result.Detail.Identity.Address,
			IdType:        int64(result.Detail.Identity.IdType),
			IdNo:          result.Detail.Identity.IdNo,
			FrontImage:    result.Detail.Identity.FrontImage,
			BackImage:     result.Detail.Identity.BackImage,
			HandheldImage: result.Detail.Identity.HandheldImage,
			KycLevel:      int64(result.Detail.Identity.KycLevel),
			VerifyStatus:  int64(result.Detail.Identity.VerifyStatus),
			RejectReason:  result.Detail.Identity.RejectReason,
			SubmitTime:    result.Detail.Identity.SubmitTime,
			VerifyTime:    result.Detail.Identity.VerifyTime,
			VerifyBy:      result.Detail.Identity.VerifyBy,
			CreateTimes:   result.Detail.Identity.CreateTimes,
			UpdateTimes:   result.Detail.Identity.UpdateTimes,
		},
		Security: types.UserSecurity{
			Id:              result.Detail.Security.Id,
			TenantId:        result.Detail.Security.TenantId,
			UserId:          result.Detail.Security.UserId,
			PayPasswordHash: result.Detail.Security.PayPasswordHash,
			GoogleSecret:    result.Detail.Security.GoogleSecret,
			GoogleEnabled:   result.Detail.Security.GoogleEnabled,
			LoginErrorCount: result.Detail.Security.LoginErrorCount,
			PayErrorCount:   result.Detail.Security.PayErrorCount,
			LockUntil:       result.Detail.Security.LockUntil,
			RiskLevel:       int64(result.Detail.Security.RiskLevel),
			CreateTimes:     result.Detail.Security.CreateTimes,
			UpdateTimes:     result.Detail.Security.UpdateTimes,
		},
	}

	// Map banks
	detail.Banks = make([]types.UserBankItem, len(result.Detail.Banks))
	for i, bank := range result.Detail.Banks {
		detail.Banks[i] = types.UserBankItem{
			Id:          bank.Id,
			TenantId:    bank.TenantId,
			UserId:      bank.UserId,
			BankName:    bank.BankName,
			BankCode:    bank.BankCode,
			AccountName: bank.AccountName,
			AccountNo:   bank.AccountNo,
			BranchName:  bank.BranchName,
			CountryCode: bank.CountryCode,
			IsDefault:   bank.IsDefault,
			Status:      int64(bank.Status),
			CreateTimes: bank.CreateTimes,
			UpdateTimes: bank.UpdateTimes,
		}
	}

	return &types.UpdateUserBaseResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Detail: detail,
	}, nil
}
