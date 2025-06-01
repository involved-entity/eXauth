package usersTest

import (
	"auth/api/users"
	"auth/tests/utils"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

var usersClient users.UsersClient
var usersJWT string
var usersUserData utils.UserData

func TestMain(m *testing.M) {
	cl, conn := utils.InitTest(users.NewUsersClient)
	usersClient = cl

	exitCode := m.Run()

	utils.ExitTest(conn, exitCode)
}

func TestGetMe(t *testing.T) {
	usersUserData, usersJWT = utils.GetAuthorizedUser()

	tt := []map[string]any{
		{
			"token":   "invalid",
			"success": false,
		},
		{
			"token":   usersJWT,
			"success": true,
		},
	}

	for _, tc := range tt {
		response, err := usersClient.GetMe(context.Background(), &users.GetMeRequest{
			Token: tc["token"].(string),
		})

		if tc["success"].(bool) {
			require.NoError(t, err)
			require.True(t, response.User.Email == usersUserData.Email)
			require.True(t, response.User.Id == int64(usersUserData.ID))
			require.True(t, response.User.Username == usersUserData.Username)
			require.True(t, response.User.IsVerified)
		} else {
			require.Error(t, err)
		}
	}
}

func TestUpdateMe(t *testing.T) {
	tt := []map[string]any{
		{
			"token":    "invalid",
			"username": "inv",
			"email":    "invalid.com",
			"success":  false,
		},
		{
			"token":    usersJWT,
			"username": "inv",
			"email":    "invalid.com",
			"success":  false,
		},
		{
			"token":    usersJWT,
			"username": "inv",
			"email":    "invalid.com",
			"success":  false,
		},
		{
			"token":    usersJWT,
			"username": usersUserData.Username,
			"email":    "example123@gmail.com",
			"success":  true,
		},
	}

	for _, tc := range tt {
		response, err := usersClient.UpdateMe(context.Background(), &users.UpdateMeRequest{
			Token:    tc["token"].(string),
			Username: tc["username"].(string),
			Email:    tc["email"].(string),
		})

		if tc["success"].(bool) {
			require.NoError(t, err)
			usersUserData.Username = response.User.Username
			usersUserData.Email = response.User.Email
		} else {
			require.Error(t, err)
		}
	}

	tt = []map[string]any{
		{
			"token":        usersJWT,
			"password":     "invalid",
			"new_password": "invalid",
			"success":      false,
		},
		{
			"token":        usersJWT,
			"password":     "invalid",
			"new_password": usersUserData.Password,
			"success":      false,
		},
		{
			"token":        usersJWT,
			"password":     usersUserData.Password,
			"new_password": "inv",
			"success":      false,
		},
		{
			"token":        usersJWT,
			"password":     usersUserData.Password,
			"new_password": usersUserData.Password,
			"success":      true,
		},
	}

	for _, tc := range tt {
		_, err := usersClient.UpdateMe(context.Background(), &users.UpdateMeRequest{
			Token:       tc["token"].(string),
			Password:    tc["password"].(string),
			NewPassword: tc["new_password"].(string),
		})

		if tc["success"].(bool) {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
	}
}

func TestGetUser(t *testing.T) {
	_, jwt := utils.GetAuthorizedUser()

	tt := []map[string]any{
		{
			"token":   "invalid",
			"id":      int64(usersUserData.ID),
			"success": false,
		},
		{
			"token":   jwt,
			"id":      int64(0),
			"success": false,
		},
		{
			"token":   jwt,
			"id":      int64(usersUserData.ID),
			"success": true,
		},
	}

	for _, tc := range tt {
		response, err := usersClient.GetUser(context.Background(), &users.GetUserRequest{
			Token: tc["token"].(string),
			Id:    tc["id"].(int64),
		})

		if tc["success"].(bool) {
			require.NoError(t, err)
			require.True(t, response.User.Email == usersUserData.Email)
			require.True(t, response.User.Id == int64(usersUserData.ID))
			require.True(t, response.User.Username == usersUserData.Username)
			require.True(t, response.User.IsVerified)
		} else {
			require.Error(t, err)
		}
	}
}
