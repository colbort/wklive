// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/rpc/system"

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
	// uidAny := l.ctx.Value("uid")
	// uid, ok := uidAny.(int64)
	// if !ok || uid <= 0 {
	// 	return nil, errorx.Wrap(nil, "未登录或登录已失效")
	// }
	out, err := l.svcCtx.SystemCli.GetProfile(l.ctx, &system.ProfileReq{Uid: 1})
	if err != nil {
		return nil, err
	}
	resp = new(types.ProfileResp)
	resp.Code = 200
	resp.Msg = "获取成功"
	_ = copier.Copy(resp, out)
	return resp, nil
}
