package utils

import (
	"log"
	"os"
	"testing"

	conf "auth/internal/pkg/config"
	"auth/internal/pkg/database"
	"auth/internal/pkg/redis"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Response struct {
	Msg string
}

type UserData struct {
	ID       int
	Username string
	Password string
	Email    string
}

func InitTest[T interface{}](clientConstructor func(grpc.ClientConnInterface) T) (T, *grpc.ClientConn) {
	conf.MustLoad()
	config := conf.GetConfig()
	conn, err := grpc.NewClient("localhost:"+config.GRPC.Port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	redis.Init(config.Redis.Address, config.Redis.Password, config.Redis.DB)

	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}

	database.Init(config.DSN)

	client := clientConstructor(conn)

	return client, conn
}

func ExitTest(con *grpc.ClientConn, exitCode int) {
	con.Close()

	os.Exit(exitCode)
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
