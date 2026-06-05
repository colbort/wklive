package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type VerificationCodeRecordDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVerificationCodeRecordDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerificationCodeRecordDetailLogic {
	return &VerificationCodeRecordDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 验证码发送记录详情
func (l *VerificationCodeRecordDetailLogic) VerificationCodeRecordDetail(in *system.VerificationCodeRecordDetailReq) (*system.VerificationCodeRecordDetailResp, error) {
	if in.Id <= 0 {
		return &system.VerificationCodeRecordDetailResp{
			Base: helper.GetErrResp(i18n.InvalidRequest, i18n.Translate(i18n.InvalidRequest, l.ctx)),
		}, nil
	}

	record, err := l.svcCtx.VerificationCodeRecordModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if err == models.ErrNotFound {
			return &system.VerificationCodeRecordDetailResp{
				Base: helper.GetErrResp(i18n.CodeNotFound, i18n.Translate(i18n.CodeNotFound, l.ctx)),
			}, nil
		}
		return nil, err
	}

	return &system.VerificationCodeRecordDetailResp{
		Base: helper.OkResp(),
		Data: verificationCodeRecordItem(record),
	}, nil
}
