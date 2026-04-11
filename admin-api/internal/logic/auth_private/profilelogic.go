// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth_private

import (
	"context"
	"fmt"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/system"

	"github.com/jinzhu/copier"
	"github.com/zeromicro/go-zero/core/logx"
)

type ProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProfileLogic {
	return &ProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProfileLogic) Profile(req *types.ProfileReq) (resp *types.ProfileResp, err error) {
	uid, err := utils.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.Translate(i18n.InternalServerError, l.ctx), err)
	}
	out, err := l.svcCtx.SystemCli.GetProfile(l.ctx, &system.ProfileReq{Uid: uid})
	if err != nil {
		return nil, err
	}
	resp = new(types.ProfileResp)
	resp.Code = i18n.OK
	resp.Msg = i18n.Translate(i18n.OK, l.ctx)
	_ = copier.Copy(&resp.Data.User, &out.User)
	_ = copier.Copy(&resp.Data.Menus, &out.Menus)
	_ = copier.Copy(&resp.Data.Perms, &out.Perms)
	return resp, nil
}
