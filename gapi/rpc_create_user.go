package gapi

import (
	"context"
	db "github.com/gaggudeep/bank_go/db/sqlc"
	"github.com/gaggudeep/bank_go/pb"
	"github.com/gaggudeep/bank_go/util"
	"github.com/gaggudeep/bank_go/validator"
	"github.com/gaggudeep/bank_go/worker"
	"github.com/hibiken/asynq"
	"github.com/lib/pq"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	violations := validateCreateUserRequest(req)
	if violations != nil {
		return nil, createInvalidRequestError(violations)
	}

	hashedPwd, err := util.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
	}

	arg := db.CreateUserTxParams{
		CreateUserParams: db.CreateUserParams{
			Username:       req.GetUsername(),
			HashedPassword: hashedPwd,
			Name:           req.GetName(),
			Email:          req.GetEmail(),
		},
		AfterCreate: func(user db.User) error {
			payload := &worker.PayloadUserCreationSuccessEmail{
				Username: user.Username,
			}
			opt := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(1 * time.Second),
				asynq.Queue(worker.QueueCritical),
			}

			return server.taskDistributor.DistributeSendUserCreationSuccessEmailTask(ctx, payload, opt...)
		},
	}

	createUserTxRes, err := server.store.CreateUserTx(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code.Name() == "unique_violation" {
				return nil, status.Errorf(codes.AlreadyExists, "duplicate unique field(s): %v", err)
			}
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	resp := &pb.CreateUserResponse{
		User: convertUser(&createUserTxRes.User),
	}

	return resp, nil
}

func validateCreateUserRequest(req *pb.CreateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validator.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, createFieldViolation("username", err.Error()))
	}
	if err := validator.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, createFieldViolation("password", err.Error()))
	}
	if err := validator.ValidateName(req.GetName()); err != nil {
		violations = append(violations, createFieldViolation("name", err.Error()))
	}
	if err := validator.ValidateEmail(req.GetEmail()); err != nil {
		violations = append(violations, createFieldViolation("email", err.Error()))
	}
	return
}
