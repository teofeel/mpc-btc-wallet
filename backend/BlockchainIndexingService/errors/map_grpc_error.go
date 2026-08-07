package errors

import (
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func MapToGRPCError(err error) error {
	switch {
	case errors.Is(err, ErrInvalidArgument), errors.Is(err, ErrInvalidAddress), errors.Is(err, ErrInvalidTransaction):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, ErrNotImplemented):
		return status.Error(codes.Unimplemented, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
