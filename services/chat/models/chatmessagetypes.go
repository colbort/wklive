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
	UserId     int64  `bson:"user_id,omitempty" json:"user_id,omitempty"`
	AgentId    int64  `bson:"agent_id,omitempty" json:"agent_id,omitempty"`

	SenderType int64              `bson:"sender_type,omitempty" json:"sender_type,omitempty"`
	Sender     *ChatMessageSender `bson:"sender,omitempty" json:"sender,omitempty"`

	MessageType int64  `bson:"message_type,omitempty" json:"message_type,omitempty"`
	Content     string `bson:"content,omitempty" json:"content,omitempty"`
	MediaUrl    string `bson:"media_url,omitempty" json:"media_url,omitempty"`
	MediaName   string `bson:"media_name,omitempty" json:"media_name,omitempty"`
	MediaMime   string `bson:"media_mime,omitempty" json:"media_mime,omitempty"`
	MediaSize   int64  `bson:"media_size,omitempty" json:"media_size,omitempty"`

	Status   int64          `bson:"status,omitempty" json:"status,omitempty"`
	Payload  map[string]any `bson:"payload,omitempty" json:"payload,omitempty"`
	ReadTime int64          `bson:"read_time,omitempty" json:"read_time,omitempty"`

	CreateTimes int64 `bson:"create_times,omitempty" json:"create_times,omitempty"`
	UpdateTimes int64 `bson:"update_times,omitempty" json:"update_times,omitempty"`
}

type ChatMessageSender struct {
	Id        int64  `bson:"id,omitempty" json:"id,omitempty"`
	Type      int64  `bson:"type,omitempty" json:"type,omitempty"`
	Nickname  string `bson:"nickname,omitempty" json:"nickname,omitempty"`
	AvatarUrl string `bson:"avatar_url,omitempty" json:"avatar_url,omitempty"`
}
