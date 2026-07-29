package adminlogic

import (
	"context"
	"errors"
	"strings"

	"wklive/common/pageutil"
	"wklive/proto/common"
	"wklive/proto/market"
	"wklive/services/market/internal/svc"
	"wklive/services/market/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAuthorityRegistriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAuthorityRegistriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAuthorityRegistriesLogic {
	return &ListAuthorityRegistriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListAuthorityRegistriesLogic) ListAuthorityRegistries(in *market.ListAuthorityRegistriesReq) (*market.ListAuthorityRegistriesResp, error) {
	if in == nil || in.Page == nil {
		return nil, errors.New("page is required")
	}
	snapshotKind := strings.ToUpper(strings.TrimSpace(in.SnapshotKind))
	if snapshotKind != "" && !isAuthorityKind(snapshotKind) {
		return nil, errors.New("unsupported snapshot kind")
	}
	status := int64(in.Status)
	if status != 0 &&
		status != int64(common.Enable_ENABLE_ENABLED) &&
		status != int64(common.Enable_ENABLE_DISABLED) {
		return nil, errors.New("invalid authority status")
	}
	rows, total, err := l.svcCtx.AuthorityRegistryAdminModel.FindPage(
		l.ctx,
		models.AuthorityRegistryFilter{
			Authority:    strings.ToLower(strings.TrimSpace(in.Authority)),
			ProviderCode: strings.ToUpper(strings.TrimSpace(in.ProviderCode)),
			ProducerType: strings.ToUpper(strings.TrimSpace(in.ProducerType)),
			SnapshotKind: snapshotKind,
			Status:       status,
		},
		in.Page.Cursor,
		in.Page.Limit,
	)
	if err != nil {
		return nil, err
	}
	data := make([]*market.AuthorityRegistryData, 0, len(rows))
	for _, row := range rows {
		item, convertErr := authorityRegistryProto(row)
		if convertErr != nil {
			return nil, convertErr
		}
		data = append(data, item)
	}
	lastID := int64(0)
	if len(rows) > 0 {
		lastID = rows[len(rows)-1].Id
	}
	return &market.ListAuthorityRegistriesResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(rows), total, lastID),
		Data: data,
	}, nil
}
