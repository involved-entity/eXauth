package users

import (
	"auth/api/users"
	utils "auth/internal/grpc"
	"auth/internal/pkg/database"
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type usersAPI struct {
	users.UnimplementedUsersServer
}

func Register(gRPC *grpc.Server) {
	users.RegisterUsersServer(gRPC, &usersAPI{})
}

type GetMeDTO struct {
	Token string `json:"token" validate:"required"`
}

func (s *usersAPI) GetMe(c context.Context, r *users.GetMeRequest) (*users.GetMeResponse, error) {
	dto := GetMeDTO{
		Token: r.Token,
	}

	err := utils.ValidateRequest(dto)
	if err != nil {
		return nil, err
	}

	rep := Repository{db: database.GetDB()}
	id, err := utils.GetUserIDByJWT(dto.Token)
	if err != nil {
		return nil, err
	}

	user, err := rep.GetUser(utils.UserInfo{ID: id})
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &users.GetMeResponse{
		User: &users.User{
			Id:         int64(user.ID),
			Username:   user.Username,
			Email:      user.Email,
			IsVerified: user.IsVerified,
			IsAdmin:    user.IsAdmin,
		},
	}, nil
}
