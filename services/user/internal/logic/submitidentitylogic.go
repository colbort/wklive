package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/notify"
	"wklive/common/utils"
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
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	// 获取用户信息
	tuser, err := l.svcCtx.UserModel.FindOne(l.ctx, userId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if tuser == nil {
		return &user.SubmitIdentityResp{
			Base: helper.ErrResp(i18n.UserNotFound, i18n.Translate(i18n.UserNotFound, l.ctx)),
		}, nil
	}

	now := utils.NowMillis()
	identity, err := l.svcCtx.UserIdentityModel.FindOneByTenantIdUserId(l.ctx, tuser.TenantId, userId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	kycLevel := int64(1) // 默认KYC等级为1
	if identity != nil {
		kycLevel = identity.KycLevel
	}
	// TODO: 根据提交的身份信息完善KYC等级评估逻辑

	if identity != nil {
		// 更新现有身份信息
		if in.Phone != "" {
			identity.Phone = sql.NullString{String: in.Phone, Valid: true}
		}
		if in.Email != "" {
			identity.Email = sql.NullString{String: in.Email, Valid: true}
		}
		if in.RealName != "" {
			identity.RealName = sql.NullString{String: in.RealName, Valid: true}
		}
		if in.Gender != 0 {
			identity.Gender = int64(in.Gender)
		}
		if in.Birthday != 0 {
			identity.Birthday = in.Birthday
		}
		if in.CountryCode != "" {
			identity.CountryCode = sql.NullString{String: in.CountryCode, Valid: true}
		}
		if in.Province != "" {
			identity.Province = sql.NullString{String: in.Province, Valid: true}
		}
		if in.City != "" {
			identity.City = sql.NullString{String: in.City, Valid: true}
		}
		if in.Address != "" {
			identity.Address = sql.NullString{String: in.Address, Valid: true}
		}
		if in.IdType != 0 {
			identity.IdType = int64(in.IdType)
		}
		if in.IdNo != "" {
			identity.IdNo = sql.NullString{String: in.IdNo, Valid: true}
		}
		if in.FrontImage != "" {
			identity.FrontImage = sql.NullString{String: in.FrontImage, Valid: true}
		}
		if in.BackImage != "" {
			identity.BackImage = sql.NullString{String: in.BackImage, Valid: true}
		}
		if in.HandheldImage != "" {
			identity.HandheldImage = sql.NullString{String: in.HandheldImage, Valid: true}
		}
		identity.KycLevel = kycLevel
		identity.VerifyStatus = 1 // 待审核
		identity.SubmitTime = now
		identity.UpdateTimes = now

		err = l.svcCtx.UserIdentityModel.Update(l.ctx, identity)
		if err != nil {
			return nil, err
		}
	} else {
		// 创建新的身份信息
		identity = &models.TUserIdentity{
			Id:            l.svcCtx.Node.Generate().Int64(),
			TenantId:      tuser.TenantId,
			UserId:        userId,
			Phone:         sql.NullString{String: in.Phone, Valid: in.Phone != ""},
			Email:         sql.NullString{String: in.Email, Valid: in.Email != ""},
			RealName:      sql.NullString{String: in.RealName, Valid: in.RealName != ""},
			Gender:        int64(in.Gender),
			Birthday:      in.Birthday,
			CountryCode:   sql.NullString{String: in.CountryCode, Valid: in.CountryCode != ""},
			Province:      sql.NullString{String: in.Province, Valid: in.Province != ""},
			City:          sql.NullString{String: in.City, Valid: in.City != ""},
			Address:       sql.NullString{String: in.Address, Valid: in.Address != ""},
			IdType:        int64(in.IdType),
			IdNo:          sql.NullString{String: in.IdNo, Valid: in.IdNo != ""},
			FrontImage:    sql.NullString{String: in.FrontImage, Valid: in.FrontImage != ""},
			BackImage:     sql.NullString{String: in.BackImage, Valid: in.BackImage != ""},
			HandheldImage: sql.NullString{String: in.HandheldImage, Valid: in.HandheldImage != ""},
			KycLevel:      kycLevel,
			VerifyStatus:  1, // 待审核
			SubmitTime:    now,
			CreateTimes:   now,
			UpdateTimes:   now,
		}

		_, err = l.svcCtx.UserIdentityModel.Insert(l.ctx, identity)
		if err != nil {
			return nil, err
		}
	}

	l.Logger.Infof("用户 %d 提交实名认证信息成功，状态为待审核", userId)

	event := notify.NewEvent(notify.EventTypeUserIdentitySubmit, notify.EventLevelInfo, "实名认证提交", fmt.Sprintf("用户 %d 提交实名认证", userId))
	event.Source = "user"
	event.TenantID = tuser.TenantId
	event.UserID = tuser.Id
	event.Data = map[string]any{
		"username": tuser.Username,
		"realName": in.RealName,
		"idType":   in.IdType.String(),
	}
	if err := notify.Publish(l.ctx, l.svcCtx.Redis, event); err != nil {
		l.Errorf("publish admin user notification failed, userId=%d err=%v", tuser.Id, err)
	}

	return &user.SubmitIdentityResp{
		Base: helper.OkResp(),
		Data: toUserIdentityProto(identity),
	}, nil
}
