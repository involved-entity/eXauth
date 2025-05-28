package tests

import (
	"auth/api/auth"
	conf "auth/internal/pkg/config"
	"auth/internal/pkg/redis"
	utils "auth/tests"
	"context"
	"strconv"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

var authJWT string

var authClient auth.AuthClient

var authUserData utils.UserData

func TestMain(m *testing.M) {
	cl, conn := utils.InitTest(auth.NewAuthClient)
	authClient = cl

	exitCode := m.Run()

	utils.ExitTest(conn, exitCode)
}

func TestRegister(t *testing.T) {
	authUserData.Email = gofakeit.Email()
	authUserData.Username = gofakeit.Username()
	authUserData.Password = gofakeit.Password(true, true, true, false, false, 8)

	tt := []map[string]any{
		{
			"email":    "invalid.com",
			"username": authUserData.Username,
			"password": authUserData.Password,
			"success":  false,
		},
		{
			"email":    authUserData.Email,
			"username": "inv",
			"password": authUserData.Password,
			"success":  false,
		},
		{
			"email":    authUserData.Email,
			"username": authUserData.Username,
			"password": "inv",
			"success":  false,
		},
		{
			"email":    authUserData.Email,
			"username": authUserData.Username,
			"password": authUserData.Password,
			"success":  true,
		},
	}

	for _, tc := range tt {
		response, err := authClient.Register(context.Background(), &auth.RegisterRequest{
			Email:    tc["email"].(string),
			Username: tc["username"].(string),
			Password: tc["password"].(string),
		})

		if tc["success"].(bool) {
			require.NoError(t, err)
			authUserData.ID = int(response.Id)
		} else {
			require.Error(t, err)
		}
	}
}

func TestRegenerateCode(t *testing.T) {
	tt := []map[string]any{
		{
			"id":      0,
			"email":   authUserData.Email,
			"success": false,
		},
		{
			"id":      authUserData.ID,
			"email":   "invalid.com",
			"success": false,
		},
		{
			"id":      authUserData.ID,
			"email":   authUserData.Email,
			"success": true,
		},
	}

	for _, tc := range tt {
		response, err := authClient.RegenerateCode(context.Background(), &auth.RegenerateCodeRequest{
			Id:    int64(tc["id"].(int)),
			Email: tc["email"].(string),
		})

		if tc["success"].(bool) {
			utils.AssertSuccess(t, err, response.GetMsg())
		} else {
			require.Error(t, err)
		}
	}
}

func TestActivateAccount(t *testing.T) {
	config := conf.GetConfig()
	name := config.OTP.RedisName
	redisClient := redis.GetClient()

	otp, err := redisClient.Get(context.Background(), name+":"+strconv.Itoa(authUserData.ID)).Result()

	require.NoError(t, err)

	tt := []map[string]any{
		{
			"id":      0,
			"code":    "invalid",
			"success": false,
		},
		{
			"id":      authUserData.ID,
			"code":    "invalid",
			"success": false,
		},
		{
			"id":      authUserData.ID,
			"code":    otp,
			"success": true,
		},
	}

	for _, tc := range tt {
		response, err := authClient.ActivateAccount(context.Background(), &auth.ActivateAccountRequest{
			Id:   int64(tc["id"].(int)),
			Code: tc["code"].(string),
		})

		if tc["success"].(bool) {
			utils.AssertSuccess(t, err, response.GetMsg())
		} else {
			require.Error(t, err)
		}
	}
}

func TestLogin(t *testing.T) {
	tt := []map[string]any{
		{
			"username": "inv",
			"password": authUserData.Password,
			"success":  false,
		},
		{
			"username": authUserData.Username,
			"password": "invalid",
			"success":  false,
		},
		{
			"username": authUserData.Username,
			"password": authUserData.Password,
			"success":  true,
		},
	}

	for _, tc := range tt {
		response, err := authClient.Login(context.Background(), &auth.LoginRequest{
			Username: tc["username"].(string),
			Password: tc["password"].(string),
		})

		if tc["success"].(bool) {
			require.NoError(t, err)
			require.True(t, utils.IsValidJWT(response.Token))

			authJWT = response.Token
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
			"token":   authJWT,
			"success": true,
		},
	}

	for _, tc := range tt {
		response, err := authClient.IsAdmin(context.Background(), &auth.IsAdminRequest{
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
			"username": authUserData.Username,
			"success":  true,
		},
	}

	for _, tc := range tt {
		response, err := authClient.ResetPassword(context.Background(), &auth.ResetPasswordRequest{
			Username: tc["username"].(string),
		})

		if tc["success"].(bool) {
			utils.AssertSuccess(t, err, response.GetMsg())
		} else {
			require.Error(t, err)
		}
	}
}

func TestResetPasswordConfirm(t *testing.T) {
	config := conf.GetConfig()
	name := config.ResetToken.RedisName
	redisClient := redis.GetClient()

	token, err := redisClient.Get(context.Background(), name+":"+strconv.Itoa(authUserData.ID)).Result()

	require.NoError(t, err)

	tt := []map[string]any{
		{
			"id":       0,
			"token":    "invalid",
			"password": "invalid",
			"success":  false,
		},
		{
			"id":       authUserData.ID,
			"token":    "invalid",
			"password": "invalid",
			"success":  false,
		},
		{
			"id":       authUserData.ID,
			"token":    token,
			"password": authUserData.Password,
			"success":  true,
		},
	}

	for _, tc := range tt {
		response, err := authClient.ResetPasswordConfirm(context.Background(), &auth.ResetPasswordConfirmRequest{
			Id:       int64(tc["id"].(int)),
			Password: tc["password"].(string),
			Token:    tc["token"].(string),
		})

		if tc["success"].(bool) {
			utils.AssertSuccess(t, err, response.GetMsg())
		} else {
			require.Error(t, err)
		}
	}
}
