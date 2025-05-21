package grpc

import (
	"auth/auth/auth"
	"auth/internal/database"
	"context"
	"time"

	conf "auth/internal/config"

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

	err := ValidateRequest(r, dto)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(r.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to hash password")
	}

	rep := Repository{db: database.GetDB()}
	user, err := rep.SaveUser(r.Username, r.Email, string(hashedPassword))
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, "this user already exists")
	}

	return &auth.RegisterResponse{Id: int64(user.ID)}, nil
}

func (s *serverAPI) Login(c context.Context, r *auth.LoginRequest) (*auth.LoginResponse, error) {
	dto := LoginRequestDTO{
		Username: r.Username,
		Password: r.Password,
	}

	err := ValidateRequest(r, dto)
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

	err := ValidateRequest(r, dto)
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
