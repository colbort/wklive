package logic

import (
	"context"

	"wklive/common/pageutil"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysChatMerchantListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysChatMerchantListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysChatMerchantListLogic {
	return &SysChatMerchantListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取客服商户列表
func (l *SysChatMerchantListLogic) SysChatMerchantList(in *system.SysChatMerchantListReq) (*system.SysChatMerchantListResp, error) {
	items, total, err := l.svcCtx.ChatMerchantModel.FindPage(l.ctx, models.ChatMerchantPageFilter{
		Keyword:      in.Keyword,
		Enabled:      commonStatusToModel(in.Enabled),
		MerchantName: in.MerchantName,
		MerchantCode: in.MerchantCode,
		ContactName:  in.ContactName,
		ContactPhone: in.ContactPhone,
		ContactEmail: in.ContactEmail,
	}, in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}

	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}

	data := make([]*system.SysChatMerchantItem, 0, len(items))
	for _, item := range items {
		data = append(data, sysChatMerchantToProto(item))
	}

	return &system.SysChatMerchantListResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastID),
		Data: data,
	}, nil
}
