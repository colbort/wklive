package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GuestLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGuestLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GuestLoginLogic {
	return &GuestLoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 游客登了
func (l *GuestLoginLogic) GuestLogin(in *user.GuestLoginReq) (*user.GuestLoginResp, error) {
	if in.DeviceId == "" && in.Fingerprint == "" {
		return &user.GuestLoginResp{
			Base: &common.RespBase{
				Code: 201,
				Msg:  "请确更换设备登录",
			},
		}, nil
	}
	u, err := l.svcCtx.UserModel.FindByDeviceIdOrFingerprint(l.ctx, in.DeviceId, in.Fingerprint)
	if err != nil && errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if u != nil {
		return &user.GuestLoginResp{
			Base: &common.RespBase{
				Code: 201,
				Msg:  "请确更换设备登录",
			},
		}, nil
	}

	err = l.checkGuestLimit(in.RegisterIp)
	userId := l.svcCtx.Node.Generate().Int64()
	token, err := utils.GenToken(
		l.svcCtx.Config.Jwt.AccessSecret,
		userId,
		fmt.Sprintf("Guest%d", userId),
		0,
		"",
		time.Duration(24*3600)*time.Second,
	)
	if err != nil {
		return nil, err
	}
	return &user.GuestLoginResp{
		Base: &common.RespBase{
			Code: 200,
			Msg:  "游客注册成功",
		},
		Data: &user.GuestLogin{
			Token:    token,
			Uid:      fmt.Sprintf("%d", userId),
			Username: fmt.Sprintf("Guest%d", userId),
			IsNew:    false,
		},
	}, nil
}

func (l *GuestLoginLogic) checkGuestLimit(ip string) error {
	key := fmt.Sprintf("guest:ip:%s:count", ip)

	// 自增次数
	count, err := l.svcCtx.Redis.Incr(key)
	if err != nil {
		return err
	}

	// 第一次自增，设置过期 1 小时
	if count == 1 {
		err := l.svcCtx.Redis.Expire(key, int(time.Hour.Seconds()))
		if err != nil {
			return err
		}
	}

	// 限制 2 次
	if count > 2 {
		return errors.New("注册过于频繁，请稍后再试")
	}

	return nil
}
