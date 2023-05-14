package gapi

import (
	"context"
	"database/sql"
	db "github.com/gaggudeep/bank_go/db/sqlc"
	"github.com/gaggudeep/bank_go/pb"
	"github.com/gaggudeep/bank_go/util"
	"github.com/gaggudeep/bank_go/validator"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	authPayload, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, createUnauthenticatedError(err)
	}
	if authPayload.Username != req.GetUsername() {
		return nil, status.Errorf(codes.PermissionDenied, "cannot update other user's info")
	}

	violations := validateUpdateUserRequest(req)
	if violations != nil {
		return nil, createInvalidRequestError(violations)
	}

	arg := db.UpdateUserParams{
		Username: req.GetUsername(),
		Name: sql.NullString{
			String: req.GetName(),
			Valid:  req.Name != nil,
		},
		Email: sql.NullString{
			String: req.GetEmail(),
			Valid:  req.Email != nil,
		},
	}

	if req.Password != nil {
		hashedPwd, err := util.HashPassword(req.GetPassword())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
		}
		arg.HashedPassword = sql.NullString{
			String: hashedPwd,
			Valid:  true,
		}
	}

	user, err := server.store.UpdateUser(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	resp := &pb.UpdateUserResponse{
		User: convertUser(&user),
	}

	return resp, nil
}

func validateUpdateUserRequest(req *pb.UpdateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validator.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, createFieldViolation("username", err.Error()))
	}
	if req.Password != nil {
		if err := validator.ValidatePassword(req.GetPassword()); err != nil {
			violations = append(violations, createFieldViolation("password", err.Error()))
		}
	}
	if req.Name != nil {
		if err := validator.ValidateName(req.GetName()); err != nil {
			violations = append(violations, createFieldViolation("name", err.Error()))
		}
	}
	if req.Email != nil {
		if err := validator.ValidateEmail(req.GetEmail()); err != nil {
			violations = append(violations, createFieldViolation("email", err.Error()))
		}
	}
	return
}

func createUnauthenticatedError(err error) error {
	return status.Errorf(codes.Unauthenticated, "unauthorized: %v", err)
}
