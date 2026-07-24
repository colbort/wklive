package delayqueue

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/zeromicro/go-queue/dq"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	ActionListContract   = "option_contract_list"
	ActionExpireContract = "option_contract_expire"
)

type Message struct {
	Action     string `json:"action"`
	TenantID   int64  `json:"tenantId"`
	ContractID int64  `json:"contractId"`
	DueAt      int64  `json:"dueAt"`
}

type Queue struct {
	producer dq.Producer
	consumer dq.Consumer
}

func New(enabled bool, beanstalks []dq.Beanstalk, redisConf redis.RedisConf) (*Queue, error) {
	if !enabled {
		return nil, nil
	}
	if len(beanstalks) < 2 {
		return nil, errors.New("option delay queue requires at least two beanstalkd nodes")
	}
	seen := make(map[string]struct{}, len(beanstalks))
	for _, node := range beanstalks {
		if node.Endpoint == "" || node.Tube == "" {
			return nil, errors.New("option delay queue endpoint and tube are required")
		}
		if _, ok := seen[node.Endpoint]; ok {
			return nil, fmt.Errorf("duplicate option delay queue endpoint: %s", node.Endpoint)
		}
		seen[node.Endpoint] = struct{}{}
	}
	return &Queue{producer: dq.NewProducer(beanstalks), consumer: dq.NewConsumer(dq.DqConf{
		Beanstalks: beanstalks, Redis: redisConf,
	})}, nil
}

func (q *Queue) At(message Message, at time.Time) error {
	if q == nil {
		return nil
	}
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}
	_, err = q.producer.At(body, at)
	return err
}

func (q *Queue) Consume(handler func(Message)) {
	if q == nil || handler == nil {
		return
	}
	q.consumer.Consume(func(body []byte) {
		var message Message
		if json.Unmarshal(body, &message) == nil {
			handler(message)
		}
	})
}
