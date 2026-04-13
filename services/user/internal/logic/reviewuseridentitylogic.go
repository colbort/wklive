package logic

import (
	"context"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReviewUserIdentityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReviewUserIdentityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewUserIdentityLogic {
	return &ReviewUserIdentityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 管理员审核用户实名认证
func (l *ReviewUserIdentityLogic) ReviewUserIdentity(in *user.ReviewUserIdentityReq) (*user.ReviewUserIdentityResp, error) {
	// 获取用户身份信息
	userIdentity, err := l.svcCtx.UserIdentityModel.FindOneByTenantIdUserId(l.ctx, in.TenantId, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if userIdentity == nil {
		return &user.ReviewUserIdentityResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.UserIdentityInfoNotFound, l.ctx)),
		}, nil
	}

	now := utils.NowMillis()

	// 更新审核信息
	userIdentity.VerifyStatus = int64(in.VerifyStatus)
	userIdentity.RejectReason.String = in.RejectReason
	userIdentity.RejectReason.Valid = in.RejectReason != ""
	userIdentity.VerifyTime = now
	userIdentity.VerifyBy.Int64 = in.VerifyBy
	userIdentity.VerifyBy.Valid = in.VerifyBy > 0
	userIdentity.UpdateTimes = now

	err = l.svcCtx.UserIdentityModel.Update(l.ctx, userIdentity)
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("管理员 %d 审核用户 %d 实名信息，状态：%d", in.VerifyBy, in.UserId, in.VerifyStatus)

	identityProto := &user.UserIdentity{
		Id:            userIdentity.Id,
		TenantId:      userIdentity.TenantId,
		UserId:        userIdentity.UserId,
		Phone:         userIdentity.Phone.String,
		Email:         userIdentity.Email.String,
		RealName:      userIdentity.RealName.String,
		Gender:        user.Gender(userIdentity.Gender),
		Birthday:      userIdentity.Birthday,
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

	return &user.ReviewUserIdentityResp{
		Base:     helper.OkResp(),
		Identity: identityProto,
	}, nil
}
