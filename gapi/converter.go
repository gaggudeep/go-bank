package gapi

import (
	db "github.com/gaggudeep/bank_go/db/sqlc"
	"github.com/gaggudeep/bank_go/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertUser(user *db.User) *pb.User {
	return &pb.User{
		Username:          user.Username,
		Name:              user.Name,
		Email:             user.Email,
		PasswordChangedAt: timestamppb.New(user.PasswordChangedAt),
		CreatedAt:         timestamppb.New(user.CreatedAt),
	}
}
