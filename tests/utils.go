package tests

import (
	"testing"

	conf "auth/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Response struct {
	Msg string
}

func AssertSuccess(t *testing.T, err error, message string) {
	require.NoError(t, err)
	assert.Equal(t, message, "success")
}

func IsValidJWT(tokenString string) bool {
	config := conf.GetConfig()

	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(config.JWT.SECRET), nil
	})

	return err == nil
}
