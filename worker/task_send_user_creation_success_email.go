package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const (
	taskSendUserCreationSuccessEmail = "task:send_user_creation_success_email"
	Type                             = "type"
	LoggerMaxRetry                   = "max_retry"
	Payload                          = "payload"
	Email                            = "email"
	LoggerQueue                      = "queue"
)

type PayloadUserCreationSuccessEmail struct {
	Username string `json:"username"`
}

func (distributor *RedisTakDistributor) DistributeSendUserCreationSuccessEmailTask(
	ctx context.Context, payload *PayloadUserCreationSuccessEmail, opt ...asynq.Option) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %v", err)
	}

	task := asynq.NewTask(taskSendUserCreationSuccessEmail, jsonPayload, opt...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueu task: %v", err)
	}

	log.Info().
		Str(Type, task.Type()).
		Bytes(Payload, task.Payload()).
		Str(LoggerQueue, info.Queue).
		Int(LoggerMaxRetry, info.MaxRetry).
		Msg("enqueued task")

	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskSendUserCreationSuccessEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadUserCreationSuccessEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w\noriginal error: %w", asynq.SkipRetry, err)
	}

	user, err := processor.store.GetUser(ctx, payload.Username)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}

	log.Info().
		Str(Type, task.Type()).
		Bytes(Payload, task.Payload()).
		Str(Email, user.Email).
		Msg("processed task")

	return nil
}
