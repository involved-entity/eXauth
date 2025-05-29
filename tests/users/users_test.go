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

	response, err := usersClient.GetMe(context.Background(), &users.GetMeRequest{
		Token: usersJWT,
	})

	require.NoError(t, err)
	require.True(t, response.User.Email == usersUserData.Email)
	require.True(t, response.User.Id == int64(usersUserData.ID))
	require.True(t, response.User.Username == usersUserData.Username)
	require.True(t, response.User.IsVerified)
}
