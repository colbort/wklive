package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysChatMerchantDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysChatMerchantDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysChatMerchantDetailLogic {
	return &SysChatMerchantDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取客服商户详情
func (l *SysChatMerchantDetailLogic) SysChatMerchantDetail(in *system.SysChatMerchantDetailReq) (*system.SysChatMerchantDetailResp, error) {
	if in.Id == nil && in.MerchantCode == nil {
		return &system.SysChatMerchantDetailResp{
			Base: helper.FailWithCode(i18n.ParamError),
		}, nil
	}

	var (
		result *models.SysChatMerchant
		err    error
	)
	if in.Id != nil {
		result, err = l.svcCtx.ChatMerchantModel.FindOne(l.ctx, *in.Id)
	} else {
		result, err = l.svcCtx.ChatMerchantModel.FindOneByMerchantCode(l.ctx, *in.MerchantCode)
	}
	if errors.Is(err, models.ErrNotFound) || result == nil {
		return &system.SysChatMerchantDetailResp{
			Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx)),
		}, nil
	}
	if err != nil {
		return nil, err
	}

	return &system.SysChatMerchantDetailResp{
		Base: helper.OkResp(),
		Data: sysChatMerchantToProto(result),
	}, nil
}
