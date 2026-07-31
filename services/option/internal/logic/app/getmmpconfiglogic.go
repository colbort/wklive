package applogic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMMPConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMMPConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMMPConfigLogic {
	return &GetMMPConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询当前用户指定做市报价组的 MMP 状态
func (l *GetMMPConfigLogic) GetMMPConfig(in *option.GetMMPConfigReq) (*option.GetMMPConfigResp, error) {
	tenantID, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	userID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	groupCode, valid := NormalizeMMPGroup(in.GroupCode)
	if tenantID <= 0 || userID <= 0 || in.ContractId <= 0 || !valid {
		return &option.GetMMPConfigResp{
			Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx)),
		}, nil
	}
	item, err := l.svcCtx.OptionMmpConfigModel.FindOneByTenantIdUserIdContractIdGroupCode(
		l.ctx, tenantID, userID, in.ContractId, groupCode,
	)
	if errors.Is(err, models.ErrNotFound) {
		return &option.GetMMPConfigResp{
			Base: helper.ErrResp(i18n.OperationNotAllowed, controlReasonMMPNotConfigured),
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &option.GetMMPConfigResp{Base: helper.OkResp(), Data: helpers.ToMMPConfigProto(item)}, nil
}
