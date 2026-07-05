package logic

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"wklive/common/helper"
	"wklive/common/utils"

	"wklive/proto/chat"
	ih "wklive/services/chat/internal/helper"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var safeHexColorRE = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$`)

type UpdateChatConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateChatConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChatConfigLogic {
	return &UpdateChatConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新chat-ui配置
func (l *UpdateChatConfigLogic) UpdateChatConfig(in *chat.UpdateChatConfigReq) (*chat.AdminChatConfigResp, error) {
	merchantID, err := ih.MerchantIDFromMetadata(l.ctx)
	if err != nil {
		return &chat.AdminChatConfigResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	data, err := l.svcCtx.ChatMerchantInfoModel.FindOneByMerchantId(l.ctx, merchantID)
	if err == models.ErrNotFound {
		return &chat.AdminChatConfigResp{Base: helper.ErrResp(404, "chat merchant config not found")}, nil
	}
	if err != nil {
		return &chat.AdminChatConfigResp{Base: helper.ErrResp(500, err.Error())}, nil
	}

	changed := false
	if title := strings.TrimSpace(in.GetTitle()); title != "" {
		data.Title = title
		changed = true
	}
	if in.GetUiConfig() != nil {
		if err := validateChatThemeConfig(in.GetUiConfig()); err != nil {
			return &chat.AdminChatConfigResp{Base: helper.ErrResp(400, err.Error())}, nil
		}
		data.UiConfig = protoMessageToNullString(mergeChatThemeConfig(ih.NullStringToChatThemeConfig(data.UiConfig), in.GetUiConfig()))
		changed = true
	}
	if in.GetFeatureConfig() != nil {
		data.FeatureConfig = protoMessageToNullStringWithDefaults(in.GetFeatureConfig())
		changed = true
	}
	if changed {
		data.UpdateTimes = utils.NowMillis()
		if err := l.svcCtx.ChatMerchantInfoModel.Update(l.ctx, data); err != nil {
			return &chat.AdminChatConfigResp{Base: helper.ErrResp(500, err.Error())}, nil
		}
	}
	return &chat.AdminChatConfigResp{Base: helper.OkResp(), Data: ih.ToProtoMerchant(data)}, nil
}

func validateChatThemeConfig(cfg *chat.ChatThemeConfig) error {
	if err := validateColorField("background_color", cfg.GetBackgroundColor()); err != nil {
		return err
	}
	if err := validateColorField("primary_color", cfg.GetPrimaryColor()); err != nil {
		return err
	}
	if err := validateColorField("notice_bar_color", cfg.GetNoticeBarColor()); err != nil {
		return err
	}
	if err := validateColorField("notice_text_color", cfg.GetNoticeTextColor()); err != nil {
		return err
	}
	if err := validateColorField("agent_bubble_color", cfg.GetAgentBubbleColor()); err != nil {
		return err
	}
	return validateColorField("user_bubble_color", cfg.GetUserBubbleColor())
}

func validateColorField(field, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if !safeHexColorRE.MatchString(value) {
		return fmt.Errorf("%s must be a hex color like #FFF, #FFFFFF, or #FFFFFFFF", field)
	}
	return nil
}

func mergeChatThemeConfig(old, patch *chat.ChatThemeConfig) *chat.ChatThemeConfig {
	if old == nil {
		old = &chat.ChatThemeConfig{}
	}
	if patch.GetBackgroundColor() != "" {
		old.BackgroundColor = patch.GetBackgroundColor()
	}
	if patch.GetPrimaryColor() != "" {
		old.PrimaryColor = patch.GetPrimaryColor()
	}
	if patch.GetNoticeBarColor() != "" {
		old.NoticeBarColor = patch.GetNoticeBarColor()
	}
	if patch.GetNoticeTextColor() != "" {
		old.NoticeTextColor = patch.GetNoticeTextColor()
	}
	if patch.GetAgentBubbleColor() != "" {
		old.AgentBubbleColor = patch.GetAgentBubbleColor()
	}
	if patch.GetUserBubbleColor() != "" {
		old.UserBubbleColor = patch.GetUserBubbleColor()
	}
	return old
}

func protoMessageToNullString(msg proto.Message) sql.NullString {
	if msg == nil {
		return sql.NullString{}
	}
	bs, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(msg)
	if err != nil || strings.TrimSpace(string(bs)) == "{}" {
		return sql.NullString{}
	}
	return sql.NullString{String: string(bs), Valid: true}
}

func protoMessageToNullStringWithDefaults(msg proto.Message) sql.NullString {
	if msg == nil {
		return sql.NullString{}
	}
	bs, err := protojson.MarshalOptions{UseProtoNames: false, EmitUnpopulated: true}.Marshal(msg)
	if err != nil {
		return sql.NullString{}
	}
	return sql.NullString{String: string(bs), Valid: true}
}
