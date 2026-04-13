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
	identity, err := l.svcCtx.UserIdentityModel.FindOneByTenantIdUserId(l.ctx, in.TenantId, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if identity == nil {
		return &user.ReviewUserIdentityResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.UserIdentityInfoNotFound, l.ctx)),
		}, nil
	}

	now := utils.NowMillis()

	// 更新审核信息
	identity.VerifyStatus = int64(in.VerifyStatus)
	identity.RejectReason.String = in.RejectReason
	identity.RejectReason.Valid = in.RejectReason != ""
	identity.VerifyTime = now
	identity.VerifyBy.Int64 = in.VerifyBy
	identity.VerifyBy.Valid = in.VerifyBy > 0
	identity.UpdateTimes = now

	err = l.svcCtx.UserIdentityModel.Update(l.ctx, identity)
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("管理员 %d 审核用户 %d 实名信息，状态：%d", in.VerifyBy, in.UserId, in.VerifyStatus)

	return &user.ReviewUserIdentityResp{
		Base:     helper.OkResp(),
		Identity: toUserIdentityProto(identity),
	}, nil
}
