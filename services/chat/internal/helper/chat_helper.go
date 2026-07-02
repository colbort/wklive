package helper

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"wklive/common/helper"
	"wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/common"

	"google.golang.org/protobuf/types/known/structpb"
)

const (
	DefaultAgentMaxSessionCount = 10
	DefaultDeviceID             = ""
)

var sequence uint64

func MerchantIDFromMetadata(ctx context.Context) (int64, error) {
	merchantID, err := utils.GetMerchantIdFromMd(ctx)
	if err != nil || merchantID <= 0 {
		return 0, errors.New("merchant_id is required")
	}
	return merchantID, nil
}

func ChatAppIdentityFromMetadata(ctx context.Context) (int64, int64, error) {
	merchantID, err := MerchantIDFromMetadata(ctx)
	if err != nil {
		return 0, 0, err
	}
	userID, err := utils.GetUserIdFromMd(ctx)
	if err != nil || userID == 0 {
		return 0, 0, errors.New("user_id is required")
	}
	return merchantID, userID, nil
}

func OffsetBase(cursor, limit int64, size int, total int64) *common.RespBase {
	if cursor < 0 {
		cursor = 0
	}
	if limit <= 0 {
		limit = pageutil.NormalizeLimit(limit)
	}
	nextCursor := cursor + int64(size)
	hasNext := nextCursor < total
	prevCursor := cursor - limit
	if prevCursor < 0 {
		prevCursor = 0
	}
	return helper.OkWithOthers(total, hasNext, cursor > 0, nextCursor, prevCursor)
}

func ValidateSessionKey(merchantID int64, sessionNo string) error {
	if merchantID <= 0 {
		return fmt.Errorf("merchant_id is required")
	}
	if strings.TrimSpace(sessionNo) == "" {
		return fmt.Errorf("session_no is required")
	}
	return nil
}

func TrimSummary(content, mediaName, mediaURL string) string {
	summary := strings.TrimSpace(content)
	if summary == "" {
		summary = strings.TrimSpace(mediaName)
	}
	if summary == "" {
		summary = strings.TrimSpace(mediaURL)
	}
	if len([]rune(summary)) <= 200 {
		return summary
	}
	return string([]rune(summary)[:200])
}

func StructToNullString(st *structpb.Struct) sql.NullString {
	if st == nil {
		return sql.NullString{}
	}
	bs, err := json.Marshal(st.AsMap())
	if err != nil {
		return sql.NullString{}
	}
	return sql.NullString{String: string(bs), Valid: true}
}

func NullStringToStruct(ns sql.NullString) *structpb.Struct {
	if !ns.Valid || strings.TrimSpace(ns.String) == "" {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(ns.String), &m); err != nil {
		return nil
	}
	st, err := structpb.NewStruct(m)
	if err != nil {
		return nil
	}
	return st
}

func MapToStruct(m map[string]any) *structpb.Struct {
	if len(m) == 0 {
		return nil
	}
	st, err := structpb.NewStruct(m)
	if err != nil {
		return nil
	}
	return st
}
