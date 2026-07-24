package secondsqueue

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/zeromicro/go-queue/dq"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	ActionActivate = "activate"
	ActionSettle   = "settle"
)

type Message struct {
	Action   string `json:"action"`
	TenantID int64  `json:"tenantId"`
	OrderID  int64  `json:"orderId"`
	Version  int64  `json:"version"`
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
		return nil, errors.New("seconds delay queue requires at least two beanstalkd nodes")
	}
	seen := make(map[string]struct{}, len(beanstalks))
	for _, node := range beanstalks {
		if node.Endpoint == "" || node.Tube == "" {
			return nil, errors.New("seconds delay queue beanstalk endpoint and tube are required")
		}
		if _, exists := seen[node.Endpoint]; exists {
			return nil, fmt.Errorf("duplicate seconds delay queue endpoint: %s", node.Endpoint)
		}
		seen[node.Endpoint] = struct{}{}
	}
	return &Queue{
		producer: dq.NewProducer(beanstalks),
		consumer: dq.NewConsumer(dq.DqConf{Beanstalks: beanstalks, Redis: redisConf}),
	}, nil
}

func (q *Queue) Delay(message Message, delay time.Duration) error {
	if q == nil {
		return nil
	}
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}
	_, err = q.producer.Delay(body, delay)
	return err
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
		if err := json.Unmarshal(body, &message); err != nil {
			return
		}
		handler(message)
	})
}
