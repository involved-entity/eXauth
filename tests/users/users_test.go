package usersTest

import (
	"auth/api/users"
	utils "auth/tests"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

var usersClient users.UsersClient

func TestMain(m *testing.M) {
	cl, conn := utils.InitTest(users.NewUsersClient)
	usersClient = cl

	exitCode := m.Run()

	utils.ExitTest(conn, exitCode)
}

func TestGetMe(t *testing.T) {
	authJWT := utils.GetAuthJWT()

	_, err := usersClient.GetMe(context.Background(), &users.GetMeRequest{
		Token: authJWT,
	})

	require.NoError(t, err)
}
