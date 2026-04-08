// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/common"
	"wklive/proto/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserIdentitiesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListUserIdentitiesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserIdentitiesLogic {
	return &ListUserIdentitiesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListUserIdentitiesLogic) ListUserIdentities(req *types.ListUserIdentitiesReq) (resp *types.ListUserIdentitiesResp, err error) {
	result, err := l.svcCtx.UserCli.ListUserIdentities(l.ctx, &user.ListUserIdentitiesReq{
		Page: &common.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
		TenantId:     req.TenantId,
		TenantCode:   req.TenantCode,
		Keyword:      req.Keyword,
		UserId:       req.UserId,
		UserNo:       req.UserNo,
		Username:     req.Username,
		Phone:        req.Phone,
		Email:        req.Email,
		RealName:     req.RealName,
		VerifyStatus: user.VerifyStatus(req.VerifyStatus),
		KycLevel:     user.KycLevel(req.KycLevel),
		IdType:       user.IdType(req.IdType),
	})
	if err != nil {
		return nil, err
	}

	data := make([]types.UserIdentityItem, len(result.List))
	for i, item := range result.List {
		data[i] = types.UserIdentityItem{
			UserId:       item.UserId,
			UserNo:       item.UserNo,
			Username:     item.Username,
			Phone:        item.Phone,
			Email:        item.Email,
			RealName:     item.RealName,
			VerifyStatus: int64(item.VerifyStatus),
			KycLevel:     int64(item.KycLevel),
			IdType:       int64(item.IdType),
			IdNo:         item.IdNo,
			SubmitTime:   item.SubmitTime,
			VerifyTime:   item.VerifyTime,
		}
	}

	return &types.ListUserIdentitiesResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: data,
	}, nil
}
