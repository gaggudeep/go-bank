package gapi

import (
	"context"
	db "github.com/gaggudeep/bank_go/db/sqlc"
	"github.com/gaggudeep/bank_go/pb"
	"github.com/gaggudeep/bank_go/util"
	"github.com/gaggudeep/bank_go/validator"
	"github.com/lib/pq"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreatUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	violations := validateCreateUserRequest(req)
	if violations != nil {
		return nil, createInvalidRequestError(violations)
	}

	hashedPwd, err := util.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %s", err)
	}

	arg := db.CreateUserParams{
		Username:       req.GetUsername(),
		HashedPassword: hashedPwd,
		Name:           req.GetName(),
		Email:          req.GetEmail(),
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code.Name() == "unqiue_violation" {
				return nil, status.Errorf(codes.AlreadyExists, "username already exists: %s", err)
			}
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %s", err)

	}

	resp := &pb.CreateUserResponse{
		User: convertUser(&user),
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
