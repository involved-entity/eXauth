package grpc

import (
	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ValidateRequest(dto interface{}) error {
	validate := validator.New()
	err := validate.Struct(dto)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return nil
}
