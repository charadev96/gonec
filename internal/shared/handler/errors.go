package handler

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ErrInternal(err error) error {
	return status.Error(codes.Internal, err.Error())
}

func ErrArg(err error) error {
	return status.Error(codes.InvalidArgument, err.Error())
}
