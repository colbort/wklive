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

type ReviewUserIdentityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReviewUserIdentityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewUserIdentityLogic {
	return &ReviewUserIdentityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReviewUserIdentityLogic) ReviewUserIdentity(req *types.ReviewUserIdentityReq) (resp *types.ReviewUserIdentityResp, err error) {
	result, err := l.svcCtx.UserCli.ReviewUserIdentity(l.ctx, &user.ReviewUserIdentityReq{
		TenantId:     req.TenantId,
		UserId:       req.UserId,
		VerifyStatus: user.VerifyStatus(req.VerifyStatus),
		RejectReason: req.RejectReason,
		VerifyBy:     req.VerifyBy,
	})
	if err != nil {
		return nil, err
	}

	return &types.ReviewUserIdentityResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Identity: types.UserIdentity{
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
			CreateTime:    result.Identity.CreateTime,
			UpdateTime:    result.Identity.UpdateTime,
		},
	}, nil
}
