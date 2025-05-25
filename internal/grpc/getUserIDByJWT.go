package grpc

import (
	conf "auth/internal/pkg/config"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
