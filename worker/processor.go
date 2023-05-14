package worker

import (
	"context"
	db "github.com/gaggudeep/bank_go/db/sqlc"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)

type TaskProcessor interface {
	Start() error
	ProcessTaskSendUserCreationSuccessEmail(context.Context, *asynq.Task) error
}

type RedisTaskProcessor struct {
	server *asynq.Server
	store  db.Store
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store) TaskProcessor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  5,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Error().
					Err(err).
					Str(Type, task.Type()).
					Bytes(Payload, task.Payload()).
					Msg("task processing failed")
			}),
			Logger: NewLogger(),
		},
	)

	return &RedisTaskProcessor{
		server,
		store,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(taskSendUserCreationSuccessEmail, processor.ProcessTaskSendUserCreationSuccessEmail)

	return processor.server.Start(mux)
}
