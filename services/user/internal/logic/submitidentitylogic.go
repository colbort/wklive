package logic

import (
	"context"
	"database/sql"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
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
	// 获取用户信息
	tuser, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if tuser == nil {
		return &user.SubmitIdentityResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.UserNotFound, l.ctx)),
		}, nil
	}

	now := utils.NowMillis()
	identity, err := l.svcCtx.UserIdentityModel.FindOneByTenantIdUserId(l.ctx, tuser.TenantId, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if identity != nil {
		// 更新现有身份信息
		identity.Phone = sql.NullString{String: in.Phone, Valid: in.Phone != ""}
		identity.Email = sql.NullString{String: in.Email, Valid: in.Email != ""}
		identity.RealName = sql.NullString{String: in.RealName, Valid: in.RealName != ""}
		identity.Gender = int64(in.Gender)
		identity.Birthday = in.Birthday
		identity.CountryCode = sql.NullString{String: in.CountryCode, Valid: in.CountryCode != ""}
		identity.Province = sql.NullString{String: in.Province, Valid: in.Province != ""}
		identity.City = sql.NullString{String: in.City, Valid: in.City != ""}
		identity.Address = sql.NullString{String: in.Address, Valid: in.Address != ""}
		identity.IdType = int64(in.IdType)
		identity.IdNo = sql.NullString{String: in.IdNo, Valid: in.IdNo != ""}
		identity.FrontImage = sql.NullString{String: in.FrontImage, Valid: in.FrontImage != ""}
		identity.BackImage = sql.NullString{String: in.BackImage, Valid: in.BackImage != ""}
		identity.HandheldImage = sql.NullString{String: in.HandheldImage, Valid: in.HandheldImage != ""}
		identity.KycLevel = int64(in.KycLevel)
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
			UserId:        in.UserId,
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
			KycLevel:      int64(in.KycLevel),
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

	l.Logger.Infof("用户 %d 提交实名认证信息成功，状态为待审核", in.UserId)

	return &user.SubmitIdentityResp{
		Base:     helper.OkResp(),
		Identity: toUserIdentityProto(identity),
	}, nil
}
