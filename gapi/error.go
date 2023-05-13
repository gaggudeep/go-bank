package gapi

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func createFieldViolation(field string, err string) *errdetails.BadRequest_FieldViolation {
	return &errdetails.BadRequest_FieldViolation{
		Field:       field,
		Description: err,
	}
}

func createInvalidRequestError(violations []*errdetails.BadRequest_FieldViolation) error {
	badReq := &errdetails.BadRequest{
		FieldViolations: violations,
	}
	statusInvalid := status.New(codes.InvalidArgument, "invalid request")

	statusDetails, err := statusInvalid.WithDetails(badReq)
	if err != nil {
		return statusInvalid.Err()
	}
	return statusDetails.Err()
}
