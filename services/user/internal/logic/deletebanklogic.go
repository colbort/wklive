package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteBankLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteBankLogic {
	return &DeleteBankLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除银行卡
func (l *DeleteBankLogic) DeleteBank(in *user.DeleteBankReq) (*user.AppCommonResp, error) {
	// 获取银行卡信息
	bank, err := l.svcCtx.UserBankModel.FindOne(l.ctx, in.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if bank == nil {
		return &user.AppCommonResp{
			Base: helper.GetErrResp(404, "银行卡不存在"),
		}, nil
	}

	// 验证银行卡是否属于该用户
	if bank.UserId != in.UserId {
		return &user.AppCommonResp{
			Base: helper.GetErrResp(403, "无权删除此银行卡"),
		}, nil
	}

	// 逻辑删除：设置status为禁用或标记删除
	// 这里我们实现为实际删除，如果需要软删除可以修改为更新status
	err = l.svcCtx.UserBankModel.Delete(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("用户 %d 删除银行卡 %d 成功", in.UserId, in.Id)

	return &user.AppCommonResp{
		Base: helper.OkResp(),
	}, nil
}
