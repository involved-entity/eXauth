package grpc

import (
	"auth/api/auth"
	"auth/internal/database"
	"context"
	"time"

	conf "auth/internal/pkg/config"

	"github.com/golang-jwt/jwt/v5"
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

type serverAPI struct {
	auth.UnimplementedAuthServer
}

func GetUserIDByJWT(JWT string) (int, error) {
	config := conf.GetConfig()
	parsedToken, err := jwt.Parse(JWT, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JWT.SECRET), nil
	})

	if err != nil {
		return 0, status.Error(codes.Unauthenticated, "invalid token")
	}

	return int(parsedToken.Claims.(jwt.MapClaims)["sub"].(map[string]interface{})["id"].(float64)), nil
}

func Register(gRPC *grpc.Server) {
	auth.RegisterAuthServer(gRPC, &serverAPI{})
}

func (s *serverAPI) Register(c context.Context, r *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	dto := RegisterRequestDTO{
		Email:    r.Email,
		Username: r.Username,
		Password: r.Password,
	}

	err := ValidateRequest(dto)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := GenerateHashedPassword(r.Password)
	if err != nil {
		return nil, err
	}

	rep := Repository{db: database.GetDB()}
	user, err := rep.SaveUser(r.Username, r.Email, string(hashedPassword))
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, "this user already exists")
	}

	if err := CreateAndSendToken(user.ID, user.Email); err != nil {
		return nil, err
	}

	return &auth.RegisterResponse{Id: int64(user.ID)}, nil
}

func (s *serverAPI) Login(c context.Context, r *auth.LoginRequest) (*auth.LoginResponse, error) {
	dto := LoginRequestDTO{
		Username: r.Username,
		Password: r.Password,
	}

	err := ValidateRequest(dto)
	if err != nil {
		return nil, err
	}

	rep := Repository{db: database.GetDB()}
	user, err := rep.GetUser(UserInfo{Username: dto.Username})
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(dto.Password)); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid password")
	}

	config := conf.GetConfig()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": JWTData{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
		"exp": time.Now().Add(time.Minute * time.Duration(config.JWT.JWT_TTL)).Unix(),
	})

	tokenString, err := token.SignedString([]byte(config.JWT.SECRET))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create token")
	}

	return &auth.LoginResponse{Token: tokenString}, nil
}

func (s *serverAPI) IsAdmin(c context.Context, r *auth.IsAdminRequest) (*auth.IsAdminResponse, error) {
	dto := IsAdminRequestDTO{
		Token: r.Token,
	}

	err := ValidateRequest(dto)
	if err != nil {
		return nil, err
	}

	rep := Repository{db: database.GetDB()}
	userID, err := GetUserIDByJWT(dto.Token)
	if err != nil {
		return nil, err
	}

	user, err := rep.GetUser(UserInfo{ID: userID})
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &auth.IsAdminResponse{IsAdmin: user.IsAdmin}, nil
}

func (s *serverAPI) RegenerateCode(c context.Context, r *auth.RegenerateCodeRequest) (*auth.RegenerateCodeResponse, error) {
	dto := RegenerateCodeDTO{ID: int(r.Id), Email: r.Email}
	if err := ValidateRequest(dto); err != nil {
		return nil, err
	}
	if err := CreateAndSendToken(uint(dto.ID), dto.Email); err != nil {
		return nil, err
	}
	return &auth.RegenerateCodeResponse{Msg: "success"}, nil
}

func (s *serverAPI) ActivateAccount(c context.Context, r *auth.ActivateAccountRequest) (*auth.ActivateAccountResponse, error) {
	dto := ActivateAccountDTO{ID: int(r.Id), Code: r.Code}
	if err := ValidateRequest(dto); err != nil {
		return nil, err
	}

	config := conf.GetConfig()
	if err := CheckRedisToken(dto.ID, dto.Code, config.OTP.RedisName); err != nil {
		return nil, err
	}

	rep := Repository{db: database.GetDB(), UserID: int(dto.ID)}
	if err := rep.VerificateUser(); err != nil {
		return nil, status.Error(codes.Internal, "failed to activate account")
	}
	return &auth.ActivateAccountResponse{Msg: "success"}, nil
}

func (s *serverAPI) ResetPassword(c context.Context, r *auth.ResetPasswordRequest) (*auth.ResetPasswordResponse, error) {
	dto := ResetPasswordDTO{Username: r.Username}
	if err := ValidateRequest(dto); err != nil {
		return nil, err
	}

	rep := Repository{db: database.GetDB()}
	user, err := rep.GetUser(UserInfo{Username: dto.Username})
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	if err := CreateAndSendResetPasswordLink(user.ID, user.Email); err != nil {
		return nil, err
	}

	return &auth.ResetPasswordResponse{Msg: "success"}, nil
}

func (s *serverAPI) ResetPasswordConfirm(c context.Context, r *auth.ResetPasswordConfirmRequest) (*auth.ResetPasswordConfirmResponse, error) {
	dto := ResetPasswordConfirmDTO{ID: int(r.Id), Token: r.Token, Password: r.Password}
	if err := ValidateRequest(dto); err != nil {
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

	rep := Repository{db: database.GetDB(), UserID: dto.ID}
	if err := rep.ChangeUserPassword(hashedPassword); err != nil {
		return nil, status.Error(codes.Internal, "failed to reset password")
	}
	return &auth.ResetPasswordConfirmResponse{Msg: "success"}, nil
}
