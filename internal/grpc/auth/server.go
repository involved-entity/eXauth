package auth

import (
	"auth/api/auth"
	"auth/internal/pkg/database"
	"context"

	conf "auth/internal/pkg/config"

	utils "auth/internal/grpc"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

type JWTData struct {
	ID       uint   `json:"id" validate:"required,gt=0"`
	Username string `json:"username" validate:"required,min=5,max=16"`
	Email    string `json:"email" validate:"required,email"`
}

type RegisterRequestDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=5,max=16"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}

type LoginRequestDTO struct {
	Username string `json:"username" validate:"required,min=5,max=16"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}

type IsAdminRequestDTO struct {
	Token string `json:"token" validate:"required"`
}

type RegenerateCodeDTO struct {
	ID    int    `json:"id" validate:"required,gt=0"`
	Email string `json:"email" validate:"required,email"`
}

type ActivateAccountDTO struct {
	ID   int    `json:"id" validate:"required,gt=0"`
	Code string `json:"code" validate:"required"`
}

type ResetPasswordDTO struct {
	Username string `json:"username" validate:"required,min=5,max=16"`
}

type ResetPasswordConfirmDTO struct {
	ID       int    `json:"id" validate:"required,gt=0"`
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}

type authAPI struct {
	auth.UnimplementedAuthServer
}

func Register(gRPC *grpc.Server) {
	auth.RegisterAuthServer(gRPC, &authAPI{})
}

func (s *authAPI) Register(c context.Context, r *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	dto := RegisterRequestDTO{
		Email:    r.Email,
		Username: r.Username,
		Password: r.Password,
	}

	err := utils.ValidateRequest(dto)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := GenerateHashedPassword(r.Password)
	if err != nil {
		return nil, err
	}

	rep := Repository{Db: database.GetDB()}
	user, err := rep.SaveUser(r.Username, r.Email, string(hashedPassword))
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, "this user already exists")
	}

	if err := CreateAndSendToken(user.ID, user.Email); err != nil {
		return nil, err
	}

	return &auth.RegisterResponse{Id: int64(user.ID)}, nil
}

func (s *authAPI) Login(c context.Context, r *auth.LoginRequest) (*auth.LoginResponse, error) {
	dto := LoginRequestDTO{
		Username: r.Username,
		Password: r.Password,
	}

	err := utils.ValidateRequest(dto)
	if err != nil {
		return nil, err
	}

	rep := Repository{Db: database.GetDB()}
	user, err := rep.GetUser(utils.UserInfo{Username: dto.Username})
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(dto.Password)); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid password")
	}

	tokenString, err := CreateJWTToken(user)
	if err != nil {
		return nil, err
	}

	return &auth.LoginResponse{Token: tokenString}, nil
}

func (s *authAPI) IsAdmin(c context.Context, r *auth.IsAdminRequest) (*auth.IsAdminResponse, error) {
	dto := IsAdminRequestDTO{
		Token: r.Token,
	}

	err := utils.ValidateRequest(dto)
	if err != nil {
		return nil, err
	}

	rep := Repository{Db: database.GetDB()}
	userID, err := utils.GetUserIDByJWT(dto.Token)
	if err != nil {
		return nil, err
	}

	user, err := rep.GetUser(utils.UserInfo{ID: userID})
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &auth.IsAdminResponse{IsAdmin: user.IsAdmin}, nil
}

func (s *authAPI) RegenerateCode(c context.Context, r *auth.RegenerateCodeRequest) (*auth.RegenerateCodeResponse, error) {
	dto := RegenerateCodeDTO{ID: int(r.Id), Email: r.Email}
	if err := utils.ValidateRequest(dto); err != nil {
		return nil, err
	}
	if err := CreateAndSendToken(uint(dto.ID), dto.Email); err != nil {
		return nil, err
	}
	return &auth.RegenerateCodeResponse{Msg: "success"}, nil
}

func (s *authAPI) ActivateAccount(c context.Context, r *auth.ActivateAccountRequest) (*auth.ActivateAccountResponse, error) {
	dto := ActivateAccountDTO{ID: int(r.Id), Code: r.Code}
	if err := utils.ValidateRequest(dto); err != nil {
		return nil, err
	}

	config := conf.GetConfig()
	if err := CheckRedisToken(dto.ID, dto.Code, config.OTP.RedisName); err != nil {
		return nil, err
	}

	rep := Repository{Db: database.GetDB(), UserID: int(dto.ID)}
	if err := rep.VerificateUser(); err != nil {
		return nil, status.Error(codes.Internal, "failed to activate account")
	}
	return &auth.ActivateAccountResponse{Msg: "success"}, nil
}

func (s *authAPI) ResetPassword(c context.Context, r *auth.ResetPasswordRequest) (*auth.ResetPasswordResponse, error) {
	dto := ResetPasswordDTO{Username: r.Username}
	if err := utils.ValidateRequest(dto); err != nil {
		return nil, err
	}

	rep := Repository{Db: database.GetDB()}
	user, err := rep.GetUser(utils.UserInfo{Username: dto.Username})
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	if err := CreateAndSendResetPasswordLink(user.ID, user.Email); err != nil {
		return nil, err
	}

	return &auth.ResetPasswordResponse{Msg: "success"}, nil
}

func (s *authAPI) ResetPasswordConfirm(c context.Context, r *auth.ResetPasswordConfirmRequest) (*auth.ResetPasswordConfirmResponse, error) {
	dto := ResetPasswordConfirmDTO{ID: int(r.Id), Token: r.Token, Password: r.Password}
	if err := utils.ValidateRequest(dto); err != nil {
		return nil, err
	}

	config := conf.GetConfig()
	if err := CheckRedisToken(dto.ID, dto.Token, config.ResetToken.RedisName); err != nil {
		return nil, err
	}

	hashedPassword, err := GenerateHashedPassword(dto.Password)
	if err != nil {
		return nil, err
	}

	rep := Repository{Db: database.GetDB(), UserID: dto.ID}
	if err := rep.ChangeUserPassword(hashedPassword); err != nil {
		return nil, status.Error(codes.Internal, "failed to reset password")
	}
	return &auth.ResetPasswordConfirmResponse{Msg: "success"}, nil
}
