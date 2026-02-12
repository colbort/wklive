package logic

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"wklive/rpc/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type SysUserCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysUserCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysUserCreateLogic {
	return &SysUserCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SysUserCreateLogic) SysUserCreate(in *system.SysUserCreateReq) (*system.RespBase, error) {
	one, err := l.svcCtx.UserModel.FindOneByUsername(l.ctx, in.Username)
	if err != nil {
		return nil, err
	}
	if one != nil {
		return &system.RespBase{
			Code: 400,
			Msg:  "用户名已存在",
		}, nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	data := models.SysUser{
		Username:      in.Username,
		Nickname:      in.Nickname,
		Password:      string(hashedPassword),
		PermsVer:      1,
		Status:        1,
		Avatar:        "",
		GoogleSecret:  "",
		GoogleEnabled: 0,
		LastLoginIp:   sql.NullString{},
		LastLoginAt:   sql.NullTime{},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	roleIds := make([]int64, 0, len(in.RoleIds))
	seen := make(map[int64]struct{}, len(in.RoleIds))
	for _, id := range in.RoleIds {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		roleIds = append(roleIds, id)
	}

	err = l.svcCtx.UserModel.TransCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		res, err := l.svcCtx.UserModel.InsertCtx(ctx, session, &data)
		if err != nil {
			return err
		}
		userId, err := res.LastInsertId()
		if err != nil {
			return err
		}
		data.Id = userId

		if len(roleIds) == 0 {
			return nil
		}

		for _, rid := range roleIds {
			role, err := l.svcCtx.RoleModel.FindOne(ctx, rid)
			if err != nil && err != sqlc.ErrNotFound {
				return err
			}
			if role == nil {
				return fmt.Errorf("role_not_found:%d", rid)
			}

			_, err = l.svcCtx.UserRoleModel.InsertCtx(ctx, session, &models.SysUserRole{
				UserId: data.Id,
				RoleId: rid,
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		if strings.HasPrefix(err.Error(), "role_not_found:") {
			return &system.RespBase{Code: 400, Msg: "角色不存在"}, nil
		}
		return nil, err
	}
	return &system.RespBase{
		Code: 200,
		Msg:  "创建成功",
	}, nil
}
