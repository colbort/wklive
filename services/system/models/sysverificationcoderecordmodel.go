package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ SysVerificationCodeRecordModel = (*customSysVerificationCodeRecordModel)(nil)

type (
	SysVerificationCodeRecordModel interface {
		sysVerificationCodeRecordModel
	}

	customSysVerificationCodeRecordModel struct {
		*defaultSysVerificationCodeRecordModel
	}
)

func NewSysVerificationCodeRecordModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) SysVerificationCodeRecordModel {
	return &customSysVerificationCodeRecordModel{
		defaultSysVerificationCodeRecordModel: newSysVerificationCodeRecordModel(conn, c, opts...),
	}
}
