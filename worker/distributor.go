package worker

import (
	"context"
	"github.com/hibiken/asynq"
)

type TaskDistributor interface {
	DistributeSendUserCreationSuccessEmailTask(context.Context, *PayloadUserCreationSuccessEmail, ...asynq.Option) error
}

type RedisTakDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor(redisOpt asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(redisOpt)

	return &RedisTakDistributor{
		client,
	}
}
