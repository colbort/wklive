package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"wklive/common/helper"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitIdentityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubmitIdentityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitIdentityLogic {
	return &SubmitIdentityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 提交实名认证信息
func (l *SubmitIdentityLogic) SubmitIdentity(in *user.SubmitIdentityReq) (*user.SubmitIdentityResp, error) {
	// 获取用户信息
	tuser, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if tuser == nil {
		return &user.SubmitIdentityResp{
			Base: helper.GetErrResp(404, "用户不存在"),
		}, nil
	}

	now := time.Now().UnixMilli()
	userIdentity, err := l.svcCtx.UserIdentityModel.FindOneByTenantIdUserId(l.ctx, tuser.TenantId, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if userIdentity != nil {
		// 更新现有身份信息
		userIdentity.Phone = sql.NullString{String: in.Phone, Valid: in.Phone != ""}
		userIdentity.Email = sql.NullString{String: in.Email, Valid: in.Email != ""}
		userIdentity.RealName = sql.NullString{String: in.RealName, Valid: in.RealName != ""}
		userIdentity.Gender = int64(in.Gender)
		userIdentity.Birthday = sql.NullTime{
			Time:  parseDate(in.Birthday),
			Valid: in.Birthday != "",
		}
		userIdentity.CountryCode = sql.NullString{String: in.CountryCode, Valid: in.CountryCode != ""}
		userIdentity.Province = sql.NullString{String: in.Province, Valid: in.Province != ""}
		userIdentity.City = sql.NullString{String: in.City, Valid: in.City != ""}
		userIdentity.Address = sql.NullString{String: in.Address, Valid: in.Address != ""}
		userIdentity.IdType = int64(in.IdType)
		userIdentity.IdNo = sql.NullString{String: in.IdNo, Valid: in.IdNo != ""}
		userIdentity.FrontImage = sql.NullString{String: in.FrontImage, Valid: in.FrontImage != ""}
		userIdentity.BackImage = sql.NullString{String: in.BackImage, Valid: in.BackImage != ""}
		userIdentity.HandheldImage = sql.NullString{String: in.HandheldImage, Valid: in.HandheldImage != ""}
		userIdentity.KycLevel = int64(in.KycLevel)
		userIdentity.VerifyStatus = 1 // 待审核
		userIdentity.SubmitTime = now
		userIdentity.UpdateTimes = now

		err = l.svcCtx.UserIdentityModel.Update(l.ctx, userIdentity)
		if err != nil {
			return nil, err
		}
	} else {
		// 创建新的身份信息
		userIdentity = &models.TUserIdentity{
			Id:       l.svcCtx.Node.Generate().Int64(),
			TenantId: tuser.TenantId,
			UserId:   in.UserId,
			Phone:    sql.NullString{String: in.Phone, Valid: in.Phone != ""},
			Email:    sql.NullString{String: in.Email, Valid: in.Email != ""},
			RealName: sql.NullString{String: in.RealName, Valid: in.RealName != ""},
			Gender:   int64(in.Gender),
			Birthday: sql.NullTime{
				Time:  parseDate(in.Birthday),
				Valid: in.Birthday != "",
			},
			CountryCode:   sql.NullString{String: in.CountryCode, Valid: in.CountryCode != ""},
			Province:      sql.NullString{String: in.Province, Valid: in.Province != ""},
			City:          sql.NullString{String: in.City, Valid: in.City != ""},
			Address:       sql.NullString{String: in.Address, Valid: in.Address != ""},
			IdType:        int64(in.IdType),
			IdNo:          sql.NullString{String: in.IdNo, Valid: in.IdNo != ""},
			FrontImage:    sql.NullString{String: in.FrontImage, Valid: in.FrontImage != ""},
			BackImage:     sql.NullString{String: in.BackImage, Valid: in.BackImage != ""},
			HandheldImage: sql.NullString{String: in.HandheldImage, Valid: in.HandheldImage != ""},
			KycLevel:      int64(in.KycLevel),
			VerifyStatus:  1, // 待审核
			SubmitTime:    now,
			CreateTimes:   now,
			UpdateTimes:   now,
		}

		_, err = l.svcCtx.UserIdentityModel.Insert(l.ctx, userIdentity)
		if err != nil {
			return nil, err
		}
	}

	l.Logger.Infof("用户 %d 提交实名认证信息成功，状态为待审核", in.UserId)

	identityProto := &user.UserIdentity{
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

	return &user.SubmitIdentityResp{
		Base:     helper.OkResp(),
		Identity: identityProto,
	}, nil
}
