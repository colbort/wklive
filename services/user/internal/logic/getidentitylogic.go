package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetIdentityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetIdentityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetIdentityLogic {
	return &GetIdentityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取身份信息
func (l *GetIdentityLogic) GetIdentity(in *user.GetIdentityReq) (*user.GetIdentityResp, error) {
	// 获取用户信息
	tuser, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if tuser == nil {
		return &user.GetIdentityResp{
			Base: helper.GetErrResp(404, "用户不存在"),
		}, nil
	}

	// 查询用户身份信息
	userIdentity, err := l.svcCtx.UserIdentityModel.FindOneByTenantIdUserId(l.ctx, tuser.TenantId, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	identityProto := &user.UserIdentity{}
	if userIdentity != nil {
		identityProto = &user.UserIdentity{
			Id:            userIdentity.Id,
			TenantId:      userIdentity.TenantId,
			UserId:        userIdentity.UserId,
			Phone:         userIdentity.Phone.String,
			Email:         userIdentity.Email.String,
			RealName:      userIdentity.RealName.String,
			Gender:        user.Gender(userIdentity.Gender),
			Birthday:      userIdentity.Birthday.Time.Format("2006-01-02"),
			CountryCode:   userIdentity.CountryCode.String,
			Province:      userIdentity.Province.String,
			City:          userIdentity.City.String,
			Address:       userIdentity.Address.String,
			IdType:        user.IdType(userIdentity.IdType),
			IdNo:          userIdentity.IdNo.String,
			FrontImage:    userIdentity.FrontImage.String,
			BackImage:     userIdentity.BackImage.String,
			HandheldImage: userIdentity.HandheldImage.String,
			KycLevel:      user.KycLevel(userIdentity.KycLevel),
			VerifyStatus:  user.VerifyStatus(userIdentity.VerifyStatus),
			RejectReason:  userIdentity.RejectReason.String,
			SubmitTime:    userIdentity.SubmitTime,
			VerifyTime:    userIdentity.VerifyTime,
			VerifyBy:      userIdentity.VerifyBy.Int64,
			CreateTimes:   userIdentity.CreateTimes,
			UpdateTimes:   userIdentity.UpdateTimes,
		}
	}

	return &user.GetIdentityResp{
		Base:     helper.OkResp(),
		Identity: identityProto,
	}, nil
}
