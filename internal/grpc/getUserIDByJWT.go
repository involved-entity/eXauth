package grpc

import (
	conf "auth/internal/pkg/config"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func getJWTData(token string) (map[string]interface{}, error) {
	config := conf.GetConfig()
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JWT.SECRET), nil
	})

	if err != nil {
		var empty map[string]interface{}
		return empty, status.Error(codes.Unauthenticated, "invalid token")
	}

	return parsedToken.Claims.(jwt.MapClaims)["sub"].(map[string]interface{}), nil
}

func GetUserIDByJWT(JWT string) (int, error) {
	data, err := getJWTData(JWT)
	if err != nil {
		return 0, err
	}
	return int(data["id"].(float64)), nil
}

func GetUserEmailAndIDByJWT(JWT string) (string, int, error) {
	data, err := getJWTData(JWT)
	if err != nil {
		return "", 0, err
	}

	return data["email"].(string), int(data["id"].(float64)), nil
}
