package models

import (
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const chatMessageCollectionPrefix = "chat_message"

func ChatMessageCollectionName(merchantId int64) string {
	if merchantId <= 0 {
		return ""
	}
	return fmt.Sprintf("%s_%d", chatMessageCollectionPrefix, merchantId)
}

type ChatMessage struct {
	ID bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`

	MessageNo  string `bson:"message_no,omitempty" json:"message_no,omitempty"`
	SessionNo  string `bson:"session_no,omitempty" json:"session_no,omitempty"`
	MerchantId int64  `bson:"merchant_id,omitempty" json:"merchant_id,omitempty"`

	Sender   *ChatMessageUser `bson:"sender,omitempty" json:"sender,omitempty"`
	Receiver *ChatMessageUser `bson:"receiver,omitempty" json:"receiver,omitempty"`

	MessageType int64  `bson:"message_type,omitempty" json:"message_type,omitempty"`
	Content     string `bson:"content,omitempty" json:"content,omitempty"`
	Url         string `bson:"url,omitempty" json:"url,omitempty"`
	FileName    string `bson:"file_name,omitempty" json:"file_name,omitempty"`
	MimeType    string `bson:"mime_type,omitempty" json:"mime_type,omitempty"`
	FileSize    int64  `bson:"file_size,omitempty" json:"file_size,omitempty"`
	Width       int32  `bson:"width,omitempty" json:"width,omitempty"`
	Height      int32  `bson:"height,omitempty" json:"height,omitempty"`
	Duration    int32  `bson:"duration,omitempty" json:"duration,omitempty"`

	Status   int64          `bson:"status,omitempty" json:"status,omitempty"`
	Payload  map[string]any `bson:"payload,omitempty" json:"payload,omitempty"`
	ReadTime int64          `bson:"read_time,omitempty" json:"read_time,omitempty"`

	CreateTimes int64 `bson:"create_times,omitempty" json:"create_times,omitempty"`
	UpdateTimes int64 `bson:"update_times,omitempty" json:"update_times,omitempty"`
}

type ChatMessageUser struct {
	Id        int64  `bson:"id,omitempty" json:"id,omitempty"`
	Type      int64  `bson:"type,omitempty" json:"type,omitempty"`
	Nickname  string `bson:"nickname,omitempty" json:"nickname,omitempty"`
	AvatarUrl string `bson:"avatar_url,omitempty" json:"avatar_url,omitempty"`
}
