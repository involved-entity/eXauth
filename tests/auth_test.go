package tests

import (
	"auth/auth/auth"
	conf "auth/internal/config"
	"auth/internal/database"
	"auth/internal/redis"
	"context"
	"log"
	"os"
	"strconv"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var JWT string

var client auth.AuthClient

type UserData struct {
	ID       int
	Username string
	Password string
	Email    string
}

var userData UserData

func InitTest() *grpc.ClientConn {
	conf.MustLoad()
	config := conf.GetConfig()
	conn, err := grpc.NewClient("localhost:"+config.GRPC.Port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	redis.Init(config.Redis.Address, config.Redis.Password, config.Redis.DB)

	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}

	database.Init(config.DSN)

	client = auth.NewAuthClient(conn)

	return conn
}

func ExitTest(con *grpc.ClientConn, exitCode int) {
	con.Close()

	os.Exit(exitCode)
}

func TestMain(m *testing.M) {
	conn := InitTest()

	exitCode := m.Run()

	ExitTest(conn, exitCode)
}

func TestRegister(t *testing.T) {
	userData.Email = gofakeit.Email()
	userData.Username = gofakeit.Username()
	userData.Password = gofakeit.Password(true, true, true, false, false, 8)

	tt := []map[string]any{
		{
			"email":    "invalid.com",
			"username": userData.Username,
			"password": userData.Password,
			"success":  false,
		},
		{
			"email":    userData.Email,
			"username": "inv",
			"password": userData.Password,
			"success":  false,
		},
		{
			"email":    userData.Email,
			"username": userData.Username,
			"password": "inv",
			"success":  false,
		},
		{
			"email":    userData.Email,
			"username": userData.Username,
			"password": userData.Password,
			"success":  true,
		},
	}

	for _, tc := range tt {
		response, err := client.Register(context.Background(), &auth.RegisterRequest{
			Email:    tc["email"].(string),
			Username: tc["username"].(string),
			Password: tc["password"].(string),
		})

		if tc["success"].(bool) {
			require.NoError(t, err)
			userData.ID = int(response.Id)
		} else {
			require.Error(t, err)
		}
	}
}

func TestRegenerateCode(t *testing.T) {
	tt := []map[string]any{
		{
			"id":      0,
			"email":   userData.Email,
			"success": false,
		},
		{
			"id":      userData.ID,
			"email":   "invalid.com",
			"success": false,
		},
		{
			"id":      userData.ID,
			"email":   userData.Email,
			"success": true,
		},
	}

	for _, tc := range tt {
		response, err := client.RegenerateCode(context.Background(), &auth.RegenerateCodeRequest{
			Id:    int64(tc["id"].(int)),
			Email: tc["email"].(string),
		})

		if tc["success"].(bool) {
			AssertSuccess(t, err, response.GetMsg())
		} else {
			require.Error(t, err)
		}
	}
}

func TestActivateAccount(t *testing.T) {
	config := conf.GetConfig()
	name := config.OTP.RedisName
	redisClient := redis.GetClient()

	otp, err := redisClient.Get(context.Background(), name+":"+strconv.Itoa(userData.ID)).Result()

	require.NoError(t, err)

	tt := []map[string]any{
		{
			"id":      0,
			"code":    "invalid",
			"success": false,
		},
		{
			"id":      userData.ID,
			"code":    "invalid",
			"success": false,
		},
		{
			"id":      userData.ID,
			"code":    otp,
			"success": true,
		},
	}

	for _, tc := range tt {
		response, err := client.ActivateAccount(context.Background(), &auth.ActivateAccountRequest{
			Id:   int64(tc["id"].(int)),
			Code: tc["code"].(string),
		})

		if tc["success"].(bool) {
			AssertSuccess(t, err, response.GetMsg())
		} else {
			require.Error(t, err)
		}
	}
}

func TestLogin(t *testing.T) {
	tt := []map[string]any{
		{
			"username": "inv",
			"password": userData.Password,
			"success":  false,
		},
		{
			"username": userData.Username,
			"password": "invalid",
			"success":  false,
		},
		{
			"username": userData.Username,
			"password": userData.Password,
			"success":  true,
		},
	}

	for _, tc := range tt {
		response, err := client.Login(context.Background(), &auth.LoginRequest{
			Username: tc["username"].(string),
			Password: tc["password"].(string),
		})

		if tc["success"].(bool) {
			require.NoError(t, err)
			require.True(t, IsValidJWT(response.Token))

			JWT = response.Token
		} else {
			require.Error(t, err)
		}
	}
}

func TestIsAdmin(t *testing.T) {
	tt := []map[string]any{
		{
			"token":   "",
			"success": false,
		},
		{
			"token":   "invalid",
			"success": false,
		},
		{
			"token":   JWT,
			"success": true,
		},
	}

	for _, tc := range tt {
		response, err := client.IsAdmin(context.Background(), &auth.IsAdminRequest{
			Token: tc["token"].(string),
		})

		if tc["success"].(bool) {
			require.NoError(t, err)
			require.False(t, response.IsAdmin)
		} else {
			require.Error(t, err)
		}
	}
}

func TestResetPassword(t *testing.T) {
	tt := []map[string]any{
		{
			"username": "invalid",
			"success":  false,
		},
		{
			"username": userData.Username,
			"success":  true,
		},
	}

	for _, tc := range tt {
		response, err := client.ResetPassword(context.Background(), &auth.ResetPasswordRequest{
			Username: tc["username"].(string),
		})

		if tc["success"].(bool) {
			AssertSuccess(t, err, response.GetMsg())
		} else {
			require.Error(t, err)
		}
	}
}

func TestResetPasswordConfirm(t *testing.T) {
	config := conf.GetConfig()
	name := config.ResetToken.RedisName
	redisClient := redis.GetClient()

	token, err := redisClient.Get(context.Background(), name+":"+strconv.Itoa(userData.ID)).Result()

	require.NoError(t, err)

	tt := []map[string]any{
		{
			"id":       0,
			"token":    "invalid",
			"password": "invalid",
			"success":  false,
		},
		{
			"id":       userData.ID,
			"token":    "invalid",
			"password": "invalid",
			"success":  false,
		},
		{
			"id":       userData.ID,
			"token":    token,
			"password": userData.Password,
			"success":  true,
		},
	}

	for _, tc := range tt {
		response, err := client.ResetPasswordConfirm(context.Background(), &auth.ResetPasswordConfirmRequest{
			Id:       int64(tc["id"].(int)),
			Password: tc["password"].(string),
			Token:    tc["token"].(string),
		})

		if tc["success"].(bool) {
			AssertSuccess(t, err, response.GetMsg())
		} else {
			require.Error(t, err)
		}
	}
}
