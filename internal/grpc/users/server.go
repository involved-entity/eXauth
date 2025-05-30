package users

import (
	"auth/api/users"
	utils "auth/internal/grpc"
	"auth/internal/pkg/database"
	"context"

	"golang.org/x/crypto/bcrypt"
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

type UpdateMeDTO struct {
	Token       string `json:"token" validate:"required"`
	Username    string `json:"username" validate:"omitempty,min=5,max=16"`
	Email       string `json:"email" validate:"omitempty,email"`
	Password    string `json:"password" validate:"omitempty,min=8,max=64"`
	NewPassword string `json:"new_password" validate:"omitempty,min=8,max=64"`
}

func (s *usersAPI) GetMe(c context.Context, r *users.GetMeRequest) (*users.GetMeResponse, error) {
	dto := GetMeDTO{
		Token: r.Token,
	}

	err := utils.ValidateRequest(dto)
	if err != nil {
		return nil, err
	}

	rep := Repository{Db: database.GetDB()}
	id, err := utils.GetUserIDByJWT(dto.Token)
	if err != nil {
		return nil, err
	}

	user, err := rep.GetUser(utils.UserInfo{ID: id, VerificateNotRequired: true})
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

func (s *usersAPI) UpdateMe(c context.Context, r *users.UpdateMeRequest) (*users.UpdateMeResponse, error) {
	dto := UpdateMeDTO{
		Token:       r.Token,
		Username:    r.Username,
		Email:       r.Email,
		Password:    r.Password,
		NewPassword: r.NewPassword,
	}

	err := utils.ValidateRequest(dto)
	if err != nil {
		return nil, err
	}

	rep := Repository{Db: database.GetDB()}
	email, id, err := utils.GetUserEmailAndIDByJWT(dto.Token)
	if err != nil {
		return nil, err
	}

	rep.UserID = id
	var newPassword string

	if dto.NewPassword != "" {
		user, err := rep.GetUser(utils.UserInfo{ID: id})
		if err != nil {
			return nil, err
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(dto.Password)); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid password")
		}

		newPassword = dto.NewPassword
	} else {
		newPassword = ""
	}

	user, err := rep.UpdateAccount(email, newPassword)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &users.UpdateMeResponse{
		User: &users.User{
			Id:         int64(user.ID),
			Username:   user.Username,
			Email:      user.Email,
			IsVerified: user.IsVerified,
			IsAdmin:    user.IsAdmin,
		},
	}, nil
}
